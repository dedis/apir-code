package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	configEnvKey      = "VPIR_CONFIG"
	defaultConfigFile = "../config.toml"

	// this key is used only for evaluation purposes
	dbPRGkey = "asuperstrong16db" // 16 bytes key for AES-128
)

func main() {
	sid := readServerID()
	logFile := flag.String("logFile", "", "write log to file instead of stdout/stderr")
	scheme := flag.String("scheme", "", "scheme to use: pir-classic, pir-merkle")
	elemBitSize := flag.Int("elemBitSize", -1, "bit size of element, in which block lengtht is specified")
	dbLen := flag.Int("dbLen", -1, "DB length in bits")
	nRows := flag.Int("nRows", -1, "number of rows in the DB representation")
	blockLen := flag.Int("blockLen", -1, "block size for DB")
	numServers := flag.Int("servers", 2, "num servers to use")

	flag.Parse()

	// write either to stdout or to logfile
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Server %v] ", sid))
	if len(*logFile) > 0 {
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatal("Could not open file: ", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("flags:", sid, *logFile, *scheme, *dbLen, *elemBitSize, *nRows, *blockLen)

	// configs
	configPath := os.Getenv(configEnvKey)
	if configPath == "" {
		configPath = defaultConfigFile
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("could not load the server config file: %v", err)
	}
	addr := config.Addresses[sid]

	// run server with TLS
	cfg := &tls.Config{
		Certificates: []tls.Certificate{utils.ServerCertificates[sid]},
		ClientAuth:   tls.NoClientCert,
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.MaxSendMsgSize(1024*1024*1024),
		grpc.Creds(credentials.NewTLS(cfg)),
	)

	// initialize DB PRG
	prgKey := new(utils.PRGKey)
	copy(prgKey[:], []byte(dbPRGkey))
	dbPRG := utils.NewPRG(prgKey)

	// Find the total number of blocks in the db
	numBlocks := *dbLen
	if (*scheme)[:3] != "cmp" {
		numBlocks = *dbLen / ((*elemBitSize) * (*blockLen))
	}
	// matrix db
	if *nRows != 1 {
		utils.IncreaseToNextSquare(&numBlocks)
		*nRows = int(math.Sqrt(float64(numBlocks)))
	}

	// initialize db
	var dbBytes *database.Bytes
	var dbFSS *database.DB
	switch *scheme {
	case "pir-classic":
		dbBytes = database.CreateRandomBytes(dbPRG, *nRows, 1, *blockLen)
	case "pir-merkle":
		dbBytes = database.CreateRandomMerkle(dbPRG, *nRows, 1, *blockLen)
	case "fss-classic", "fss-auth":
		numIdenfitiers := 100000
		dbFSS, err = database.CreateRandomKeysDB(dbPRG, numIdenfitiers)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("unknow scheme: " + string(*scheme))
	}

	// GC after db creation
	runtime.GC()

	// select correct server
	var s server.Server
	switch *scheme {
	case "pir-classic", "pir-merkle":
		if *numServers == 2 {
			s = server.NewPIRTwo(dbBytes)
			log.Println("using DPF with numServers:", numServers)
		} else {
			s = server.NewPIR(dbBytes)
		}
	case "fss-classic":
		s = server.NewPredicatePIR(dbFSS, byte(sid))
	case "fss-auth":
		s = server.NewPredicateAPIR(dbFSS, byte(sid))
	default:
		log.Fatal("unknow scheme for server: " + string(*scheme))
	}

	// start server
	proto.RegisterVPIRServer(rpcServer, &vpirServer{
		Server: s,
		scheme: *scheme,
	})
	log.Printf("is listening at %s", addr)

	// listen signals from os
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error, 1)

	go func() {
		log.Println("starting grpc server")
		if err := rpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// start HTTP server for tests
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal("impossible to parse addr for HTTP server")
	}
	h := func(w http.ResponseWriter, _ *http.Request) {
		sigCh <- os.Interrupt
	}
	httpAddr := fmt.Sprintf("%s:%s", host, "8080")
	srv := &http.Server{Addr: httpAddr}
	http.HandleFunc("/", h)
	go func() {
		srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		log.Fatalf("failed to serve: %v", err)
	case <-sigCh:
		rpcServer.GracefulStop()
		lis.Close()
		srv.Shutdown(context.Background())
		log.Println("clean shutdown of server done")
	}
}

// vpirServer is used to implement VPIR Server protocol.
type vpirServer struct {
	proto.UnimplementedVPIRServer
	Server server.Server

	scheme string
	cores  int
}

func (s *vpirServer) DatabaseInfo(ctx context.Context, r *proto.DatabaseInfoRequest) (
	*proto.DatabaseInfoResponse, error) {
	log.Print("got databaseInfo request")

	dbInfo := s.Server.DBInfo()

	if s.scheme[:3] == "fss" {
		return &proto.DatabaseInfoResponse{NumColumns: uint32(dbInfo.NumColumns)}, nil
	}

	resp := &proto.DatabaseInfoResponse{
		NumRows:     uint32(dbInfo.NumRows),
		NumColumns:  uint32(dbInfo.NumColumns),
		BlockLength: uint32(dbInfo.BlockSize),
		PirType:     dbInfo.PIRType,
		Root:        dbInfo.Root,
		ProofLen:    uint32(dbInfo.ProofLen),
	}

	return resp, nil
}

func (s *vpirServer) Query(ctx context.Context, qr *proto.QueryRequest) (
	*proto.QueryResponse, error) {
	log.Print("got query request")

	a, err := s.Server.AnswerBytes(qr.GetQuery())
	if err != nil {
		return nil, err
	}
	answerLen := len(a)
	log.Printf("stats,%d", answerLen)

	return &proto.QueryResponse{Answer: a}, nil
}

func readServerID() int {
	file, err := os.Open("sid")
	if err != nil {
		log.Fatal(err)
	}

	var sid int

	_, err = fmt.Fscanf(file, "%d\n", &sid) // give a pattern to scan
	if err != nil {
		log.Fatal(err)
	}

	return sid
}
