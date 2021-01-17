package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/si-co/vpir-code/lib/utils"

	db "github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"google.golang.org/grpc"
)

func main() {
	// flags
	sid := flag.Int("id", -1, "Server ID")
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
	vpirServer := &vpirServer{
		Server: server.NewITServer(db.CreateAsciiVector()),
	}
	proto.RegisterVPIRServer(rpcServer, vpirServer)
	log.Printf("Server %d is listening at %s", *sid, addr)

	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// vpirServer is used to implement VPIR Server protocol.
type vpirServer struct {
	proto.UnimplementedVPIRServer
	server.Server // TODO: create a general server
}

func (s *vpirServer) DatabaseInfo(ctx context.Context, r *proto.DatabaseInfoRequest) (
	*proto.DatabaseInfoResponse, error) {

	// send block length back, implement the logic in db
	blockLength = 16

	return &proto.DatabaseInfoResponse{BlockLength: blockLength}, nil
}

func (s *vpirServer) Query(ctx context.Context, qr *proto.Request) (
	*proto.Response, error) {

	query := make([]*big.Int, len(qr.Query))
	for i, v := range qr.Query {
		var ok bool
		query[i], ok = big.NewInt(0).SetString(v, 10)
		if !ok {
			log.Fatalf("Could not convert string %v to big int", v)
		}
	}

	a := s.Answer(query)
	return &proto.Response{Answer: a.String()}, nil
}
