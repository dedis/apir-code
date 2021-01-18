package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"google.golang.org/grpc"
)

func main() {
	// set logs
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Client] "))

	// flags
	idPtr := flag.String("id", "", "id for which key should be retrieved")
	flag.Parse()

	// configs
	config, err := utils.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("could not load the config file: %v", err)
	}
	addresses, err := utils.ServerAddresses(config)
	if err != nil {
		log.Fatalf("could not parse servers addresses: %v", err)
	}

	// New random generator
	var key utils.PRGKey
	_, err = io.ReadFull(rand.Reader, key[:])
	if err != nil {
		log.Fatalf("PRG initialization error: %v", err)
	}
	prg := utils.NewPRG(&key)

	// start client and initialize top-level Context
	c := client.NewDPF(prg)
	fmt.Println(c)
	ctx := context.Background()

	// get db info
	dbInfo := runDBInfoRequest(ctx, addresses)
	fmt.Println(dbInfo)

	// get id and compute corresponding hash
	id := *idPtr
	if id == "" {
		log.Fatal("id not provided")
	}
	idHash := utils.HashToIndex(id, constants.DBLength)

	// query for given idHash
	fmt.Println(idHash)
}

func runDBInfoRequest(ctx context.Context, addresses []string) int {
	subCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan int, len(addresses))
	for _, a := range addresses {
		wg.Add(1)
		go func(addr string) {
			resCh <- dbInfo(subCtx, addr)
			wg.Done()
		}(a)
	}
	wg.Wait()
	close(resCh)

	// check if db info are all equal before returning
	dbInfo := make([]int, 0)
	for i := range resCh {
		dbInfo = append(dbInfo, i)
	}
	for i := range dbInfo {
		if dbInfo[0] != dbInfo[i] {
			log.Fatal("got different database info from servers")
		}
	}

	return dbInfo[0]

}

func dbInfo(ctx context.Context, address string) int {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewVPIRClient(conn)
	q := &proto.DatabaseInfoRequest{}
	answer, err := c.DatabaseInfo(ctx, q)
	if err != nil {
		log.Fatalf("could not send database info request: %v", err)
	}

	blockLength := answer.GetBlockLength()

	return int(blockLength)
}

func runQueries(ctx context.Context, queries [][]*big.Int, addrs []string) []*big.Int {
	if len(addrs) != len(queries) {
		log.Fatal("Queries and server addresses length mismatch")
	}

	subCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan *big.Int, len(queries))
	for i := 0; i < len(queries); i++ {
		wg.Add(1)
		go func(j int) {
			resCh <- query(subCtx, addrs[j], queries[j])
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(resCh)

	return aggregateResults(resCh)
}

func query(ctx context.Context, address string, query []*big.Int) *big.Int {
	conn, err := grpc.Dial(address, grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewVPIRClient(conn)
	q := &proto.Request{Query: convertToString(query)}
	answer, err := c.Query(ctx, q)
	if err != nil {
		log.Fatalf("could not query: %v", err)
	}

	num, ok := big.NewInt(0).SetString(answer.GetAnswer(), 10)
	if !ok {
		log.Fatal("Could not convert answer to big int")
	}

	return num
}

func convertToString(query []*big.Int) []string {
	q := make([]string, len(query))
	for i, v := range query {
		q[i] = v.String()
	}
	return q
}

func aggregateResults(ch <-chan *big.Int) []*big.Int {
	q := make([]*big.Int, 0)
	for v := range ch {
		q = append(q, v)
	}
	return q
}
