package main

import (
	"log"
	"net"
	"os"

	pb "github.com/si-co/vpir-code/proto"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedVPIRServer
}

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
	s := grpc.NewServer()
	pb.RegisterVPIRServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return nil
}
