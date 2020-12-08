package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ncw/gmp"
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

	// Contact the servers and print out its response.
	output := ""
	for i := 0; i < 136; i++ {
		queries := c.Query(i, len(addresses))
		answers := runQueries(queries, addresses)
		result, err := c.Reconstruct(answers)
		if err != nil {
			log.Fatalf("Failed reconstructing %v with error: %v", i, err)
		}
		output += result.String()
	}
	b, err := utils.BitStringToBytes(output)
	if err != nil {
		log.Fatalf("Could not convert bit string to bytes: %v", err)
	}
	fmt.Println("Output: ", string(b))
}

func runQueries(queries [][]*gmp.Int, addresses []string) []*gmp.Int {
	if len(addresses) != len(queries) {
		log.Fatal("Queries and server addresses length mismatch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan *gmp.Int, len(queries))
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

func query(ctx context.Context, address string, query []*gmp.Int) *gmp.Int {
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

	num, ok := gmp.NewInt(0).SetString(answer.Answer, 10)
	if !ok {
		log.Fatal("Could not convert answer to big int")
	}

	return num
}

func convertToString(query []*gmp.Int) []string {
	q := make([]string, len(query))
	for i, v := range query {
		q[i] = v.String()
	}
	return q
}

func aggregateResults(ch <-chan *gmp.Int) []*gmp.Int {
	q := make([]*gmp.Int, 0)
	for v := range ch {
		q = append(q, v)
	}
	return q
}
