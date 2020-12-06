package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"

	db "github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"google.golang.org/grpc"
)

func main() {
	log.SetOutput(os.Stdout)

	index := flag.Int("index", -1, "Server index")
	addr := flag.String("addr", "", "Server address")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[Server %v] ", *index))

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer()
	vpirServer := &vpirServer{
		Server: server.NewITServer(db.CreateAsciiDatabase()),
	}
	proto.RegisterVPIRServer(rpcServer, vpirServer)
	log.Printf("Server %d is listening at %s", *index, *addr)

	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		fmt.Println("here")
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
