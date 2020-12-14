package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/grpc"
)

func main() {
	addresses, err := utils.LoadServerConfig("config.toml")
	if err != nil {
		log.Fatalf("Could not load the server config file: %v", err)
	}
	// New random generator
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		log.Fatalf("Could not create new XOF: %v", err)
	}
	c := client.NewITClient(xof)
	log.SetPrefix(fmt.Sprintf("[Client] "))

	// Contact the servers and print out its response.
	output := ""
	log.Printf("Start retrieving process")
	for i := 0; i < 136; i++ {
		queries := c.Query(i, len(addresses))
		log.Printf("Send %d queries for i=%d", len(queries), i)
		answers := runQueries(queries, addresses)
		log.Printf("Receive %d queries for i=%d", len(answers), i)
		result, err := c.Reconstruct(answers)
		log.Printf("Reconstructed result: %s", result.String())
		if err != nil {
			log.Fatalf("Failed reconstructing %v with error: %v", i, err)
		}
		output += result.String()
	}
	log.Printf("End retrieving process")
	b, err := utils.BitStringToBytes(output)
	if err != nil {
		log.Fatalf("Could not convert bit string to bytes: %v", err)
	}
	fmt.Println("Output: ", string(b))
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
