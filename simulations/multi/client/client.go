package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	configEnvKey = "VPIR_CONFIG"

	defaultConfigFile = "../config.toml"
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
	listenAddr string

	scheme    string
	id        string
	target    string
	fromStart int
	fromEnd   int
	and       bool
	avg       bool
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

func (lc *localClient) exec() (string, error) {
	// get and store db info.
	lc.retrieveDBInfo()

	// start correct client
	switch lc.flags.scheme {
	case "pir-classic", "pir-merkle":
		lc.vpirClient = client.NewPIR(lc.prg, lc.dbInfo)
	default:
		return "", xerrors.Errorf("wrong scheme: %s", lc.flags.scheme)
	}
	// case "pointPIR", "pointVPIR":
	// 	lc.vpirClient = client.NewPIR(lc.prg, lc.dbInfo)

	// 	// get id
	// 	if lc.flags.id == "" {
	// 		var id string
	// 		fmt.Print("please enter the id: ")
	// 		fmt.Scanln(&id)
	// 		if id == "" {
	// 			log.Fatal("id not provided")
	// 		}
	// 		lc.flags.id = id
	// 	}

	// 	// retrieve the key corresponding to the id
	// 	return lc.retrieveKeyGivenId(lc.flags.id)
	// case "complexPIR":
	// 	lc.vpirClient = client.NewPredicatePIR(lc.prg, lc.dbInfo)
	// 	out, err := lc.retrieveComplexQuery()
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	return strconv.FormatUint(uint64(out), 10), nil
	// case "complexVPIR":
	// 	lc.vpirClient = client.NewPredicateAPIR(lc.prg, lc.dbInfo)
	// 	out, err := lc.retrieveComplexQuery()
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	return strconv.FormatUint(uint64(out), 10), nil
	// default:
	// 	return "", xerrors.Errorf("wrong scheme: %s", lc.flags.scheme)
	// }

	return "", nil
}

func (lc *localClient) retrievePointPIR() {
	numTotalBlocks := lc.dbInfo.NumRows * lc.dbInfo.NumColumns
	numRetrieveBlocks := bitsToBlocks(blockSize, elemBitSize, numBitsToRetrieve)

	var startIndex int
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)

		// pick a random block index to start the retrieval
		startIndex = rand.Intn(numTotalBlocks - numRetrieveBlocks)
		for i := 0; i < numRetrieveBlocks; i++ {
			queries := c.Query(startIndex+i, len(ss))

			_, err := c.Reconstruct(answers)
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			results[j].Bandwidth[i].Reconstruct = 0
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return results
}

func (lc *localClient) connectToServers() error {
	// load servers certificates
	creds, err := utils.LoadServersCertificates()
	if err != nil {
		return xerrors.Errorf("could not load servers certificates: %v", err)
	}

	// connect to servers and store connections
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

func equalDBInfo(info []*database.Info) bool {
	for i := range info {
		if info[0].NumRows != info[i].NumRows ||
			info[0].NumColumns != info[i].NumColumns ||
			info[0].BlockSize != info[i].BlockSize {
			return false
		}
	}

	return true
}

func connectToServer(creds credentials.TransportCredentials, address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		return nil, xerrors.Errorf("did not connect to %s: %v", address, err)
	}

	log.Println("connected to server", address)

	return conn, nil
}

func parseFlags() *flags {
	f := new(flags)

	// scheme flags
	flag.StringVar(&f.scheme, "scheme", "", "scheme to use")

	// flag for point queries
	flag.StringVar(&f.id, "id", "", "id of key to retrieve")

	// flag for complex queries
	flag.StringVar(&f.target, "target", "", "target for complex query")
	flag.IntVar(&f.fromStart, "from-start", 0, "from start parameter for complex query")
	flag.IntVar(&f.fromEnd, "from-end", 0, "from end parameter for complex query")
	flag.BoolVar(&f.and, "and", false, "and clause for complex query")
	flag.BoolVar(&f.avg, "avg", false, "avg clause for complex query")

	flag.Parse()

	return f
}
