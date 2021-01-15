package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"google.golang.org/grpc"
)

func main() {
	log.SetPrefix(fmt.Sprintf("[Client] "))

	config, err := utils.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("Could not load the config file: %v", err)
	}
	addresses, err := utils.ServerAddresses(config)
	fmt.Println(addresses)

	// New random generator
	var key utils.PRGKey
	_, err = io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	prg := utils.NewPRG(&key)

	// start client
	c := client.NewDPF(prg)
	fmt.Println(c)

	// get id and compute corresponding hash
	idPtr := flag.String("id", "", "id for which key should be retrieved")
	flag.Parse()
	id := *idPtr
	idHash := utils.HashToIndex(id, constants.DBLength)

	// query for given idHash
	fmt.Println(idHash)
}

func runQueries(queries [][]*big.Int, addresses []string) []*big.Int {
	if len(addresses) != len(queries) {
		log.Fatal("Queries and server addresses length mismatch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan *big.Int, len(queries))
	for i := 0; i < len(queries); i++ {
		wg.Add(1)
		go func(j int) {
			resCh <- query(ctx, addresses[j], queries[j])
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(resCh)

	return aggregateResults(resCh)
}

func query(ctx context.Context, address string, query []*big.Int) *big.Int {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
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

	num, ok := big.NewInt(0).SetString(answer.Answer, 10)
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
