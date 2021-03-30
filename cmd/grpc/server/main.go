package main

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
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

func main() {
	// flags
	sid := flag.Int("id", -1, "Server ID")
	experiment := flag.Bool("experiment", false, "run setting for experiments")
	filesNumber := flag.Int("files", 1, "number of key files to use in db creation")
	cores := flag.Int("cores", -1, "number of cores to use")
	logFile := flag.String("log", "", "write log to file instead of stdout/stderr")
	prof := flag.Bool("prof", false, "Write CPU prof file")
	flag.Parse()

	// start profiling
	if *prof {
		utils.StartProfiling(fmt.Sprintf("server-%v.prof", *sid))
		defer utils.StopProfiling()
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

	// set stats log
	var statsLogger *log.Logger
	if *experiment {
		fileName := fmt.Sprintf("stats_server-%v.log", *sid)
		f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("could not open stats.log file: %v", err)
		}
		defer f.Close()
		statsLogger = log.New(f, "", log.Lmsgprefix)
	}

	// configs
	config, err := utils.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("could not load the server config file: %v", err)
	}
	addr := config.Addresses[*sid]

	// load the db
	fmt.Println(*filesNumber)
	//db, err := loadPgpDB(*filesNumber)
	//if err != nil {
	//log.Fatalf("impossible to construct real keys db: %v", err)
	//}
	db, err := loadPgpDBFromDisk("data/database.bin")
	if err != nil {
		log.Fatalf("impossible to construct real keys db: %v", err)
	}

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
	if *cores != -1 && *experiment {
		s = server.NewIT(db, *cores)
	} else {
		s = server.NewIT(db)
	}

	// start server
	proto.RegisterVPIRServer(rpcServer, &vpirServer{
		Server:      s,
		experiment:  *experiment,
		cores:       *cores,
		statsLogger: statsLogger,
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
	experiment  bool
	cores       int
	statsLogger *log.Logger
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
		s.statsLogger.Printf("%d,%d", s.cores, answerLen)
	}

	return &proto.QueryResponse{Answer: a}, nil
}

func loadPgpDB(filesNumber int) (*database.DB, error) {
	log.Println("Starting to read in the DB data")
	sksDir := filepath.Join("data", pgp.SksParsedFolder)
	files, err := pgp.GetAllFiles(sksDir)
	if err != nil {
		return nil, err
	}
	// take only filesNumber files
	files = files[:filesNumber]

	db, err := database.GenerateRealKeyDB(files, constants.ChunkBytesLength, false)
	if err != nil {
		return nil, err
	}
	log.Println("DB loaded with files", files)

	return db, nil
}

func loadPgpDBFromDisk(path string) (*database.DB, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db := &database.DB{}
	dec := gob.NewDecoder(f)
	if err = dec.Decode(&db); err != nil {
		return nil, err
	}

	return db, nil
}
