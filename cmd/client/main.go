package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

type localClient struct {
	connections map[string]*grpc.ClientConn
	ctx         context.Context
	callOptions []grpc.CallOption

	flags flags

	dbInfo     *database.Info
	vpirClient client.Client
}

type flags struct {
	scheme          string
	realApplication bool
	logFile         string
	profiling       bool
}

func main() {
	flags := parseFlags()

	// enable profiling
	if flags.profiling {
		utils.StartProfiling("client.prof")
		defer utils.StopProfiling()
	}

	// set logs
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Client] "))
	if len(flags.logFile) > 0 {
		f, err := os.Create(flags.logFile)
		if err != nil {
			log.Fatal("Could not open file: ", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// configs
	config, err := utils.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("could not load the config file: %v", err)
	}

	// initialize local client
	lc := &localClient{
		ctx: context.Background(),
		callOptions: []grpc.CallOption{
			grpc.UseCompressor(gzip.Name),
			grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
			grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
		},
	}

	// random generator
	prg := utils.RandomPRG()

	// load servers certificates
	creds, err := utils.LoadServersCertificates()
	if err != nil {
		log.Fatalf("could not load servers certificates: %v", err)
	}

	// connect to servers and store connections
	lc.connections = make(map[string]*grpc.ClientConn)
	for _, s := range config.Addresses {
		lc.connections[s] = connectToServer(creds, s)
		defer lc.connections[s].Close()
	}

	// get and store db info
	lc.runDBInfo()

	// start correct client
	var c client.Client
	switch flags.scheme {
	case "dpf":
		c = client.NewDPF(prg, lc.dbInfo)
	case "it":
		c = client.NewIT(prg, lc.dbInfo)
	case "pir":
		c = client.NewPIR(prg, lc.dbInfo)
	default:
		log.Fatal("undefined scheme type")
	}
	lc.vpirClient = c
	log.Printf("scheme: %s", flags.scheme)

	if !flags.realApplication {
		lc.runExperiment()
	}

	// get id and compute corresponding hash
	if flags.realApplication {
		log.Printf("running key server client")
		for {
			var id string
			fmt.Scanln(&id)
			if id == "" {
				log.Fatal("id not provided")
			}

			// compute hash key for id
			hashKey := database.HashToIndex(id, lc.dbInfo.NumRows*lc.dbInfo.NumColumns)
			log.Printf("id: %s, hashKey: %d", id, hashKey)

			// query given hash key
			queries, err := c.QueryBytes(hashKey, len(lc.connections))
			if err != nil {
				log.Fatalf("error when executing query: %v", err)
			}

			// send queries to servers
			answers := lc.runQueries(queries)

			// reconstruct block
			resultField, err := c.ReconstructBytes(answers)
			if err != nil {
				log.Fatalf("error during reconstruction: %v", err)
			}

			// return result bytes
			result := field.VectorToBytes(resultField)

			// unpad result
			result = database.UnPadBlock(result)

			// get a key from the block with the id of the search
			retrievedKey, err := pgp.RecoverKeyFromBlock(result, id)
			log.Printf("key: %#v", retrievedKey)
		}
	}
}

func (lc *localClient) runDBInfo() {
	subCtx, cancel := context.WithTimeout(lc.ctx, time.Second)
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
		IDLength:   int(answer.GetIdLength()),
		KeyLength:  int(answer.GetKeyLength()),
	}

	return dbInfo
}

func (lc *localClient) runQueries(queries [][]byte) [][]byte {
	subCtx, cancel := context.WithTimeout(lc.ctx, time.Minute)
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

	// combinate anwsers of all the servers
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

func connectToServer(creds credentials.TransportCredentials, address string) *grpc.ClientConn {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect to %s: %v", address, err)
	}

	return conn
}

func equalDBInfo(info []*database.Info) bool {
	for i := range info {
		if info[0].NumRows != info[i].NumRows ||
			info[0].NumColumns != info[i].NumColumns ||
			info[0].BlockSize != info[i].BlockSize ||
			info[0].IDLength != info[i].IDLength ||
			info[0].KeyLength != info[i].KeyLength {
			return false
		}
	}

	return true
}

func (lc *localClient) runExperiment() {
	// TODO: use separate logger experiments outputs
	// TODO: repeat the experiment multiple times using a
	// configurable parameter
	log.Printf("running client for micro-benchmarks")

	// TODO: this is for a db represented as a vector
	// find a way to unify
	numBlocks := lc.dbInfo.NumColumns * lc.dbInfo.NumRows

	// save in csv file
	// TODO: use log for this?
	fmt.Println("query,reconstruct,total_client")

	// create main monitor for CPU time
	totalTimer := monitor.NewMonitor()
	m := monitor.NewMonitor()

	for i := 0; i < numBlocks; i++ {
		// query given block
		m.Reset()
		queries, err := lc.vpirClient.QueryBytes(i, len(lc.connections))
		fmt.Printf("%.3f,", m.RecordAndReset())
		if err != nil {
			log.Fatal("error when executing query")
		}

		// send queries to servers and get answers
		answers := lc.runQueries(queries)

		// reconstruct, we don't use the actual result
		m.Reset()
		_, err = lc.vpirClient.ReconstructBytes(answers)
		fmt.Printf("%.3f,", m.RecordAndReset())
		if err != nil {
			log.Fatalf("error during reconstruction: %v", err)
		}
	}
	fmt.Printf("%.3f\n", totalTimer.Record())
}

func parseFlags() *flags {
	f := new(flags)

	flag.StringVar(&f.scheme, "scheme", "", "dpf for DPF-based and IT for information-theoretic")
	flag.StringVar(&f.logFile, "log", "", "write log to file instead of stdout/stderr")
	flag.BoolVar(&f.profiling, "prof", false, "Write pprof file")
	flag.BoolVar(&f.realApplication, "app", true, "Run key server client or client used for experiments")
	flag.Parse()

	return f
}
