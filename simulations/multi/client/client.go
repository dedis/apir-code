package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

const (
	configEnvKey = "VPIR_CONFIG"

	defaultConfigFile = "config.toml"
)

type localClient struct {
	ctx         context.Context
	callOptions []grpc.CallOption
	connections map[string]*grpc.ClientConn

	prg    *utils.PRGReader
	config *utils.Config
	flags  *flags
}

type flags struct {
	listenAddr string

	scheme    string
	id        string
	target    string
	fromStart int
	fromEnd   int
	and       bool
	avg       bool
}

func newLocalClient() *localClient {
	// initialize local client
	lc := &localClient{
		ctx: context.Background(),
		callOptions: []grpc.CallOption{
			grpc.UseCompressor(gzip.Name),
			grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
			grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
		},
		prg:   utils.RandomPRG(),
		flags: parseFlags(),
	}

	// set logs to stdout
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Client] "))

	// load configs
	configPath := os.Getenv(configEnvKey)
	if configPath == "" {
		configPath = defaultConfigFile
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("could not load the config file: %v", err)
	}
	lc.config = config

	return lc
}

func main() {
	lc := newLocalClient()

	err := lc.connectToServers()
	defer lc.closeConnections()

	if err != nil {
		log.Fatal(err)
	}
}

func (lc *localClient) connectToServers() error {
	// load servers certificates
	creds, err := utils.LoadServersCertificates()
	if err != nil {
		return xerrors.Errorf("could not load servers certificates: %v", err)
	}

	// connect to servers and store connections
	lc.connections = make(map[string]*grpc.ClientConn)
	for _, s := range lc.config.Addresses {
		conn, err := connectToServer(creds, s)
		if err != nil {
			return xerrors.Errorf("failed to connect: %v", err)
		}

		lc.connections[s] = conn
	}

	return nil
}

func (lc *localClient) closeConnections() {
	for _, conn := range lc.connections {
		err := conn.Close()
		if err != nil {
			log.Printf("failed to close conn: %v", err)
		}
	}
}

func connectToServer(creds credentials.TransportCredentials, address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		return nil, xerrors.Errorf("did not connect to %s: %v", address, err)
	}

	return conn, nil
}

func parseFlags() *flags {
	f := new(flags)

	// scheme flags
	flag.StringVar(&f.scheme, "scheme", "", "scheme to use")

	// flag for point queries
	flag.StringVar(&f.id, "id", "", "id of key to retrieve")

	// flag for complex queries
	flag.StringVar(&f.target, "target", "", "target for complex query")
	flag.IntVar(&f.fromStart, "from-start", 0, "from start parameter for complex query")
	flag.IntVar(&f.fromEnd, "from-end", 0, "from end parameter for complex query")
	flag.BoolVar(&f.and, "and", false, "and clause for complex query")
	flag.BoolVar(&f.avg, "avg", false, "avg clause for complex query")

	flag.Parse()

	return f
}
