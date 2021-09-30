package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	configEnvKey = "VPIR_CONFIG"

	defaultConfigFile = "config.toml"
)

type localClient struct {
	ctx         context.Context
	callOptions []grpc.CallOption
	connections map[string]*grpc.ClientConn

	prg        *utils.PRGReader
	config     *utils.Config
	flags      *flags
	dbInfo     *database.Info
	vpirClient client.Client
}

type flags struct {
	id        string
	profiling bool

	// only for experiments
	experiment bool
	cores      int
	scheme     string

	demo       bool
	listenAddr string
}

func newLocalClient() *localClient {
	// initialize local client
	lc := &localClient{
		ctx: context.Background(),
		callOptions: []grpc.CallOption{
			grpc.UseCompressor(gzip.Name),
			grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
			grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
		},
		prg:   utils.RandomPRG(),
		flags: parseFlags(),
	}

	// enable profiling if needed
	if lc.flags.profiling {
		utils.StartProfiling("client.prof")
		defer utils.StopProfiling()
	}

	// set logs to stdout
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Client] "))

	// load configs
	configPath := os.Getenv(configEnvKey)
	if configPath == "" {
		configPath = defaultConfigFile
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("could not load the config file: %v", err)
	}
	lc.config = config

	return lc
}

func main() {
	lc := newLocalClient()

	if lc.flags.demo {
		lc.runDemo()
		return
	} else {
		err := lc.connectToServers()
		defer lc.closeConnections()

		if err != nil {
			log.Fatal(err)
		}

		_, err = lc.exec()
		if err != nil {
			log.Fatal(err)
		}
	}

	os.Exit(0)
}

func (lc *localClient) connectToServers() error {
	// load servers certificates
	creds, err := utils.LoadServersCertificates()
	if err != nil {
		return xerrors.Errorf("could not load servers certificates: %v", err)
	}

	// connect to servers and store connections
	// TODO: move somewhere else, but mind the defer
	lc.connections = make(map[string]*grpc.ClientConn)
	for _, s := range lc.config.Addresses {
		conn, err := connectToServer(creds, s)
		if err != nil {
			return xerrors.Errorf("failed to connect: %v", err)
		}

		lc.connections[s] = conn
	}

	return nil
}

func (lc *localClient) closeConnections() {
	for _, conn := range lc.connections {
		err := conn.Close()
		if err != nil {
			log.Printf("failed to close conn: %v", err)
		}
	}
}

func (lc *localClient) exec() (string, error) {
	// get and store db info.
	// This function queries the servers for the database information.
	// In the Keyd PoC application, we will hardcode the database
	// information in the client.
	lc.retrieveDBInfo()

	// start correct client, which can be either IT or DPF.
	switch lc.flags.scheme {
	case "pointPIR", "pointVPIR":
		lc.vpirClient = client.NewPIR(lc.prg, lc.dbInfo)

		// get id
		if lc.flags.id == "" {
			var id string
			fmt.Print("please enter the id: ")
			fmt.Scanln(&id)
			if id == "" {
				log.Fatal("id not provided")
			}
			lc.flags.id = id
		}

		// retrieve the key corresponding to the id
		return lc.retrieveKeyGivenId(lc.flags.id)
	case "complexPIR":
		lc.vpirClient = client.NewPIRfss(lc.prg, lc.dbInfo)
		panic("not yet implemented")
	case "complexVPIR":
		lc.vpirClient = client.NewFSS(lc.prg, lc.dbInfo)
		panic("not yet implemented")
	default:
		return "", xerrors.Errorf("wrong scheme: %s", lc.flags.scheme)
	}
}

func (lc *localClient) retrieveKeyGivenId(id string) (string, error) {
	t := time.Now()

	// compute hash key for id
	hashKey := database.HashToIndex(id, lc.dbInfo.NumRows*lc.dbInfo.NumColumns)
	log.Printf("id: %s, hashKey: %d", id, hashKey)

	// query given hash key
	in := make([]byte, 4)
	fmt.Println("HASH KEY IN", hashKey)
	binary.BigEndian.PutUint32(in, uint32(hashKey))
	queries, err := lc.vpirClient.QueryBytes(in, len(lc.connections))
	if err != nil {
		return "", xerrors.Errorf("error when executing query: %v", err)
	}
	log.Printf("done with queries computation")

	// send queries to servers
	answers := lc.runQueries(queries)

	// reconstruct block
	resultField, err := lc.vpirClient.ReconstructBytes(answers)
	if err != nil {
		return "", xerrors.Errorf("error during reconstruction: %v", err)
	}
	log.Printf("done with block reconstruction")

	var result []byte
	if lc.flags.scheme == "it" || lc.flags.scheme == "dpf" {
		// return result bytes
		result = field.VectorToBytes(resultField)
	} else {
		result = resultField.([]byte)
	}
	// unpad result in both cases
	result = database.UnPadBlock(result)

	// get a key from the block with the id of the search
	retrievedKey, err := pgp.RecoverKeyFromBlock(result, id)
	if err != nil {
		return "", xerrors.Errorf("error retrieving key from the block: %v", err)
	}
	log.Printf("PGP key retrieved from block")

	armored, err := pgp.ArmorKey(retrievedKey)
	if err != nil {
		return "", xerrors.Errorf("error armor-encoding the key: %v", err)
	}

	fmt.Println(armored)

	elapsedTime := time.Since(t)
	if lc.flags.experiment {
		// query bw
		bw := 0
		for _, q := range queries {
			bw += len(q)
		}
		log.Printf("stats,%d,%d,%f", lc.flags.cores, bw, elapsedTime.Seconds())
	}
	fmt.Printf("Wall-clock time to retrieve the key: %v\n", elapsedTime)

	return armored, nil
}

func (lc *localClient) retrieveDBInfo() {
	subCtx, cancel := context.WithTimeout(lc.ctx, time.Hour)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan *database.Info, len(lc.connections))
	for _, conn := range lc.connections {
		wg.Add(1)
		go func(conn *grpc.ClientConn) {
			resCh <- dbInfo(subCtx, conn, lc.callOptions)
			wg.Done()
		}(conn)
	}
	wg.Wait()
	close(resCh)

	dbInfo := make([]*database.Info, 0)
	for i := range resCh {
		dbInfo = append(dbInfo, i)
	}

	// check if db info are all equal before returning
	if !equalDBInfo(dbInfo) {
		log.Fatal("got different database info from servers")
	}

	log.Printf("databaseInfo: %#v", dbInfo[0])

	lc.dbInfo = dbInfo[0]
}

func dbInfo(ctx context.Context, conn *grpc.ClientConn, opts []grpc.CallOption) *database.Info {
	c := proto.NewVPIRClient(conn)
	q := &proto.DatabaseInfoRequest{}
	answer, err := c.DatabaseInfo(ctx, q, opts...)
	if err != nil {
		log.Fatalf("could not send database info request to %s: %v",
			conn.Target(), err)
	}
	log.Printf("sent databaseInfo request to %s", conn.Target())

	dbInfo := &database.Info{
		NumRows:    int(answer.GetNumRows()),
		NumColumns: int(answer.GetNumColumns()),
		BlockSize:  int(answer.GetBlockLength()),
		PIRType:    answer.GetPirType(),
		Merkle:     &database.Merkle{Root: answer.GetRoot(), ProofLen: int(answer.GetProofLen())},
	}

	return dbInfo
}

func (lc *localClient) runQueries(queries [][]byte) [][]byte {
	subCtx, cancel := context.WithTimeout(lc.ctx, time.Hour)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan []byte, len(lc.connections))
	j := 0
	for _, conn := range lc.connections {
		wg.Add(1)
		go func(j int, conn *grpc.ClientConn) {
			resCh <- query(subCtx, conn, lc.callOptions, queries[j])
			wg.Done()
		}(j, conn)
		j++
	}
	wg.Wait()
	close(resCh)

	// combinate answers of all the servers
	q := make([][]byte, 0)
	for v := range resCh {
		q = append(q, v)
	}

	return q
}

func query(ctx context.Context, conn *grpc.ClientConn, opts []grpc.CallOption, query []byte) []byte {
	c := proto.NewVPIRClient(conn)
	q := &proto.QueryRequest{Query: query}
	answer, err := c.Query(ctx, q, opts...)
	if err != nil {
		log.Fatalf("could not query %s: %v",
			conn.Target(), err)
	}
	log.Printf("sent query to %s", conn.Target())
	log.Printf("query size in bytes %d", len(query))

	return answer.GetAnswer()
}

func connectToServer(creds credentials.TransportCredentials, address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		return nil, xerrors.Errorf("did not connect to %s: %v", address, err)
	}

	return conn, nil
}

func equalDBInfo(info []*database.Info) bool {
	for i := range info {
		if info[0].NumRows != info[i].NumRows ||
			info[0].NumColumns != info[i].NumColumns ||
			info[0].BlockSize != info[i].BlockSize {
			//info[0].IDLength != info[i].IDLength ||
			//info[0].KeyLength != info[i].KeyLength {
			return false
		}
	}

	return true
}

func parseFlags() *flags {
	f := new(flags)

	flag.BoolVar(&f.profiling, "prof", false, "write pprof file")
	flag.StringVar(&f.id, "id", "", "id of key to retrieve")
	flag.BoolVar(&f.experiment, "experiment", false, "run for experiments")
	flag.IntVar(&f.cores, "cores", -1, "num of cores used for experiment")
	flag.StringVar(&f.scheme, "scheme", "", "scheme to use: it, dpf or pit-it, pir-dpf")
	flag.BoolVar(&f.demo, "demo", false, "runs as a demo, which exposes a REST API")
	flag.StringVar(&f.listenAddr, "listen-addr", "", "demo listen address")
	flag.Parse()

	return f
}
