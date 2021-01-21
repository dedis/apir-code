package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

var creds credentials.TransportCredentials

func init() {

	// load servers certificates
	cp := x509.NewCertPool()
	for _, cert := range utils.ServerPublicKeys {
		if !cp.AppendCertsFromPEM([]byte(cert)) {
			log.Fatalf("credentials: failed to append certificates")
		}
	}
	cfg := &tls.Config{RootCAs: cp}
	// TODO: remove
	cfg.InsecureSkipVerify = true
	creds = credentials.NewTLS(cfg)
}

func main() {

	// set logs
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Client] "))

	// flags
	idPtr := flag.String("id", "", "id for which key should be retrieved")
	schemePtr := flag.String("scheme", "", "dpf for DPF-based and IT for information-theoretic")
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

	// random generator
	var key utils.PRGKey
	_, err = io.ReadFull(rand.Reader, key[:])
	if err != nil {
		log.Fatalf("PRG initialization error: %v", err)
	}
	prg := utils.NewPRG(&key)

	// initialize top level context
	ctx := context.Background()

	// get db info
	dbInfo := runDBInfoRequest(ctx, addresses)
	log.Printf("databaseInfo: %#v", dbInfo)

	// start correct client
	var c client.Client
	switch *schemePtr {
	case "dpf":
		c = client.NewDPF(prg, *dbInfo)
	case "it":
		c = client.NewITClient(prg, *dbInfo)
	default:
		log.Fatal("undefined scheme type")
	}
	log.Printf("scheme: %s", *schemePtr)

	// get id and compute corresponding hash
	id := *idPtr
	if id == "" {
		log.Fatal("id not provided")
	}
	idHash := database.HashToIndex(id, dbInfo.NumColumns*dbInfo.NumRows)
	log.Printf("id: %s, hashKey: %d", id, idHash)

	// query for given idHash
	queries, err := c.QueryBytes(idHash, len(addresses))
	if err != nil {
		log.Fatal("error when executing query")
	}

	// send queries to servers
	answers := runQueries(ctx, addresses, queries)

	res, err := c.ReconstructBytes(answers)
	if err != nil {
		log.Fatalf("error during reconstruction: %v", err)
	}
	fmt.Println("res:", res)

	// find correct key
	resultBytes := field.VectorToBytes(res)
	keyLength := dbInfo.KeyLength
	idLength := dbInfo.IDLength
	chunkLength := constants.ChunkBytesLength
	zeroSlice := make([]byte, idLength)

	// determine (id, key) length in bytes
	lastElementBytes := keyLength % chunkLength
	keyLengthWithPadding := int(math.Ceil(float64(keyLength)/float64(chunkLength))) * chunkLength
	totalLength := idLength + keyLengthWithPadding

	// parse block entries
	idKey := make(map[string]string)
	for i := 0; i < len(resultBytes)-totalLength+1; i += totalLength {
		idBytes := resultBytes[i : i+idLength]
		// test if we are in padding elements already
		if bytes.Equal(idBytes, zeroSlice) {
			break
		}
		idReconstructed := string(bytes.Trim(idBytes, "\x00"))

		keyBytes := resultBytes[i+idLength : i+idLength+keyLengthWithPadding]
		// remove padding for last element
		if lastElementBytes != 0 {
			keyBytes = append(keyBytes[:len(keyBytes)-chunkLength], keyBytes[len(keyBytes)-(lastElementBytes):]...)
		}

		// encode key
		idKey[idReconstructed] = base64.StdEncoding.EncodeToString(keyBytes)
		fmt.Println(base64.StdEncoding.EncodeToString(keyBytes))
	}
	log.Printf("key: %s", idKey[id])
}

func runDBInfoRequest(ctx context.Context, addresses []string) *database.Info {
	subCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan *database.Info, len(addresses))
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
	dbInfo := make([]*database.Info, 0)
	for i := range resCh {
		dbInfo = append(dbInfo, i)
	}
	for i := range dbInfo {
		if dbInfo[0].NumRows != dbInfo[i].NumRows ||
			dbInfo[0].NumColumns != dbInfo[i].NumColumns ||
			dbInfo[0].BlockSize != dbInfo[i].BlockSize {
			log.Fatal("got different database info from servers")
		}
	}

	return dbInfo[0]

}

func dbInfo(ctx context.Context, address string) *database.Info {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithBlock())
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
	log.Printf("sent databaseInfo request to %s", address)

	dbInfo := &database.Info{
		NumRows:    int(answer.GetNumRows()),
		NumColumns: int(answer.GetNumColumns()),
		BlockSize:  int(answer.GetBlockLength()),
		IDLength:   int(answer.GetIdLength()),
		KeyLength:  int(answer.GetKeyLength()),
	}

	return dbInfo
}

func runQueries(ctx context.Context, addrs []string, queries [][]byte) [][]byte {
	if len(addrs) != len(queries) {
		log.Fatal("Queries and server addresses length mismatch")
	}

	subCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan []byte, len(addrs))
	for i := 0; i < len(queries); i++ {
		wg.Add(1)
		go func(j int) {
			resCh <- query(subCtx, addrs[j], queries[j])
			wg.Done()
		}(i)
	}
	wg.Wait()
	close(resCh)

	// combinate ansers of all the servers
	q := make([][]byte, 0)
	for v := range resCh {
		q = append(q, v)
	}

	return q
}

func query(ctx context.Context, address string, query []byte) []byte {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := proto.NewVPIRClient(conn)
	q := &proto.QueryRequest{Query: query}
	var opts []grpc.CallOption
	opts = append(opts, grpc.UseCompressor(gzip.Name))
	answer, err := c.Query(ctx, q, opts...)
	if err != nil {
		log.Fatalf("could not query: %v", err)
	}

	return answer.GetAnswer()
}
