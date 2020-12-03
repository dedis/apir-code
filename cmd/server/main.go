package main

import (
	"context"
	db "github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/proto"
	srv "github.com/si-co/vpir-code/lib/server"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"log"
	"math/big"
	"net"
	"os"
)

func main() {
	app := &cli.App{
		Name:   "VPIR server",
		Usage:  "WIP",
		Action: runServer,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func runServer(c *cli.Context) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}
	rpcServer := grpc.NewServer()
	vpirServer := &server{
		server: srv.CreateServer(db.CreateAsciiDatabase()),
	}
	proto.RegisterVPIRServer(rpcServer, vpirServer)

	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return nil
}

type Server interface {
	Answer(q []*big.Int) *big.Int
}

// server is used to implement vpir server protocol.
type server struct {
	proto.UnimplementedVPIRServer
	server Server
}

func (s *server) Query(ctx context.Context, qr *proto.QueryRequest) (
	*proto.Answer, error) {

	query := make([]*big.Int, len(qr.Query))
	for i, v := range qr.Query {
		var ok bool
		query[i], ok = big.NewInt(0).SetString(v, 10)
		if !ok {
			log.Fatalf("Could not convert string %v to big int", v)
		}
	}

	a := s.server.Answer(query)
	return &proto.Answer{Answer: a.String()}, nil
}
