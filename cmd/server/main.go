package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/si-co/vpir-code/lib/utils"
	"log"
	"net"
	"os"

	big "github.com/ncw/gmp"
	db "github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"google.golang.org/grpc"
)

func main() {
	log.SetOutput(os.Stdout)

	sid := flag.Int("id", -1, "Server ID")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[Server %v] ", *sid))

	addrs, err := utils.LoadServerConfig("config.toml")
	if err != nil {
		log.Fatalf("Could not load the server config file: %v", err)
	}
	addr := addrs[*sid]

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer()
	vpirServer := &vpirServer{
		Server: server.NewITServer(db.CreateAsciiDatabase()),
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
	server.Server
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
