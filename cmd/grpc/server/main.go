package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"

	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	configEnvKey = "VPIR_CONFIG"
	dataEnvKey   = "VPIR_SKS_ROOT"

	defaultConfigFile = "config.toml"
	defaultSksPath    = "data"
)

func main() {
	// flags
	sid := flag.Int("id", -1, "Server ID")
	experiment := flag.Bool("experiment", false, "run setting for experiments")
	filesNumber := flag.Int("files", 1, "number of key files to use in db creation")
	cores := flag.Int("cores", -1, "number of cores to use")
	vpirScheme := flag.String("scheme", "", "vpir scheme to use: it or dpf")
	logFile := flag.String("log", "", "write log to file instead of stdout/stderr")
	prof := flag.Bool("prof", false, "Write CPU prof file")
	mprof := flag.Bool("mprof", false, "Write memory prof file")

	flag.Parse()

	// start profiling
	if *prof {
		utils.StartProfiling(fmt.Sprintf("server-%v.prof", *sid))
		defer utils.StopProfiling()
	}

	if *mprof {
		fn := fmt.Sprintf("server-%v-mem.mprof", *sid)
		defer func() {
			f, err := os.Create(fn)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Writing memory profile")
			pprof.WriteHeapProfile(f)
			f.Close()
		}()
	}

	// set logs
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Server %v] ", *sid))
	if len(*logFile) > 0 {
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatal("Could not open file: ", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// configs
	configPath := os.Getenv(configEnvKey)
	if configPath == "" {
		configPath = defaultConfigFile
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("could not load the server config file: %v", err)
	}
	addr := config.Addresses[*sid]

	// load the db
	var db *database.DB
	switch *vpirScheme {
	case "it":
		db, err = loadPgpDB(*filesNumber, true)
		if err != nil {
			log.Fatalf("impossible to construct real keys db: %v", err)
		}
	//db, err := database.LoadDB("data/db", "vpir")
	//if err != nil {
	//log.Fatalf("impossible to load real keys db: %v", err)
	//}
	case "dpf":
		// mmap db is vector for the moment
		db, err = database.LoadMMapDB("data/mmap.d/")
		if err != nil {
			log.Fatalf("impossible to construct real keys db: %v", err)
		}
	default:
		log.Fatal("unknown vpir scheme")
	}

	// GC after db creation
	runtime.GC()

	// run server with TLS
	cfg := &tls.Config{
		Certificates: []tls.Certificate{utils.ServerCertificates[*sid]},
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

	// select correct server
	var s server.Server
	switch *vpirScheme {
	case "it":
		if *cores != -1 && *experiment {
			s = server.NewIT(db, *cores)
		} else {
			s = server.NewIT(db)
		}
	case "dpf":
		if *cores != -1 && *experiment {
			s = server.NewDPF(db, *cores)
		} else {
			s = server.NewDPF(db)
		}
	default:
		log.Fatal("unknow VPIR type")
	}

	// start server
	proto.RegisterVPIRServer(rpcServer, &vpirServer{
		Server:     s,
		experiment: *experiment,
		cores:      *cores,
	})
	log.Printf("is listening at %s", addr)

	// start HTTP server for tests
	// TODO: remove this in application
	host, _ := net.SplitHostPort(addr)
	h := func(w http.ResponseWriter, _ *http.Request) {
		os.Exit(0)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", server.IP, "8080"), nil))

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

	select {
	case err := <-errCh:
		log.Fatalf("failed to serve: %v", err)
	case <-sigCh:
		rpcServer.GracefulStop()
		lis.Close()
		log.Println("clean shutdown of server done")
	}
}

// vpirServer is used to implement VPIR Server protocol.
type vpirServer struct {
	proto.UnimplementedVPIRServer
	Server server.Server // both IT and DPF-based server

	// only for experiments
	experiment bool
	cores      int
}

func (s *vpirServer) DatabaseInfo(ctx context.Context, r *proto.DatabaseInfoRequest) (
	*proto.DatabaseInfoResponse, error) {
	log.Print("got databaseInfo request")

	dbInfo := s.Server.DBInfo()
	resp := &proto.DatabaseInfoResponse{
		NumRows:     uint32(dbInfo.NumRows),
		NumColumns:  uint32(dbInfo.NumColumns),
		BlockLength: uint32(dbInfo.BlockSize),
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
	log.Printf("answer size in bytes: %d", answerLen)
	if s.experiment {
		log.Printf("stats,%d,%d", s.cores, answerLen)
	}

	return &proto.QueryResponse{Answer: a}, nil
}

func loadPgpDB(filesNumber int, rebalanced bool) (*database.DB, error) {
	log.Println("Starting to read in the DB data")

	sksDir := os.Getenv(dataEnvKey)
	if sksDir == "" {
		sksDir = filepath.Join(defaultSksPath, pgp.SksParsedFolder)
	}

	files, err := pgp.GetAllFiles(sksDir)
	if err != nil {
		return nil, err
	}
	// take only filesNumber files
	files = files[:filesNumber]

	db, err := database.GenerateRealKeyDB(files, constants.ChunkBytesLength, rebalanced)
	if err != nil {
		return nil, err
	}
	log.Println("DB loaded with files", files)

	return db, nil
}
