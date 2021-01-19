package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/utils"

	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"google.golang.org/grpc"
)

func main() {
	// flags
	sid := flag.Int("id", -1, "Server ID")
	schemePtr := flag.String("scheme", "", "dpf for DPF-based and IT for information-theoretic")
	flag.Parse()

	// set logs
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Server %v] ", *sid))

	// configs
	config, err := utils.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("could not load the server config file: %v", err)
	}
	addresses, err := utils.ServerAddresses(config)
	if err != nil {
		log.Fatalf("could not parse servers addresses: %v", err)
	}
	addr := addresses[*sid]

	// generate db
	// TODO: How do we choose dbLen (hence, nCols) ?
	dbLen := 40 * 1024 * 8
	chunkLength := constants.ChunkBytesLength // maximum numer of bytes embedded in a field elements
	nRows := 1
	nCols := dbLen / (nRows * chunkLength)
	db, err := database.GenerateKeyDB("../../data/random_id_key.csv", chunkLength, nRows, nCols)
	if err != nil {
		log.Fatalf("could not generate keys db: %v", err)
	}

	// run server with TLS
	cfg := &tls.Config{
		Certificates: []tls.Certificate{utils.ServerCertificates[*sid]},
		ClientAuth:   tls.NoClientCert,
	}
	lis, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer()

	// select correct server
	var s server.Server
	switch *schemePtr {
	case "dpf":
		s = server.NewDPF(db, byte(*sid))
	case "it":
		s = server.NewITServer(db)
	default:
		log.Fatal("undefined scheme type")
	}

	// start server
	proto.RegisterVPIRServer(rpcServer, &vpirServer{Server: s})
	log.Printf("Server %d is listening at %s", *sid, addr)

	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// vpirServer is used to implement VPIR Server protocol.
type vpirServer struct {
	proto.UnimplementedVPIRServer
	Server server.Server // both IT and DPF-based server
}

func (s *vpirServer) DatabaseInfo(ctx context.Context, r *proto.DatabaseInfoRequest) (
	*proto.DatabaseInfoResponse, error) {
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

	a, err := s.Server.AnswerBytes(qr.GetQuery())
	if err != nil {
		return nil, err
	}
	return &proto.QueryResponse{Answer: a}, nil
}
