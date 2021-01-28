package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
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

type localClient struct {
	dbInfo *database.Info
}

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

func connectToServer(address string) *grpc.ClientConn {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect to %s: %v", address, err)
	}

	return conn
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

	// random generator
	prg := utils.RandomPRG()

	// initialize top level context
	ctx := context.Background()

	// connect to servers
	connections := make(map[string]*grpc.ClientConn)
	for _, s := range config.Addresses {
		connections[s] = connectToServer(s)
		defer connections[s].Close()
	}

	// get db info
	dbInfo := runDBInfoRequest(ctx, connections)

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
	queries, err := c.QueryBytes(idHash, len(config.Addresses))
	if err != nil {
		log.Fatal("error when executing query")
	}

	// send queries to servers
	answers := runQueries(ctx, connections, queries)

	res, err := c.ReconstructBytes(answers)
	if err != nil {
		log.Fatalf("error during reconstruction: %v", err)
	}

	// retrieve bytes from field elements
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
			keyBytes = append(keyBytes[:len(keyBytes)-chunkLength],
				keyBytes[len(keyBytes)-(lastElementBytes):]...)
		}

		// encode key
		idKey[idReconstructed] = base64.StdEncoding.EncodeToString(keyBytes)
	}
	log.Printf("key: %s", idKey[id])
}

func runDBInfoRequest(ctx context.Context, cs map[string]*grpc.ClientConn) *database.Info {
	subCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan *database.Info, len(cs))
	for _, conn := range cs {
		wg.Add(1)
		go func(conn *grpc.ClientConn) {
			resCh <- dbInfo(subCtx, conn)
			wg.Done()
		}(conn)
	}
	wg.Wait()
	close(resCh)

	dbInfo := make([]*database.Info, 0)
	for i := range resCh {
		dbInfo = append(dbInfo, i)
	}

	// check if db info are all equal before returning
	if !equalDBInfo(dbInfo) {
		log.Fatal("got different database info from servers")
	}

	log.Printf("databaseInfo: %#v", dbInfo[0])

	return dbInfo[0]
}

func dbInfo(ctx context.Context, conn *grpc.ClientConn) *database.Info {
	c := proto.NewVPIRClient(conn)
	q := &proto.DatabaseInfoRequest{}
	answer, err := c.DatabaseInfo(ctx, q)
	if err != nil {
		log.Fatalf("could not send database info request to %s: %v",
			conn.Target(), err)
	}
	log.Printf("sent databaseInfo request to %s", conn.Target())

	dbInfo := &database.Info{
		NumRows:    int(answer.GetNumRows()),
		NumColumns: int(answer.GetNumColumns()),
		BlockSize:  int(answer.GetBlockLength()),
		IDLength:   int(answer.GetIdLength()),
		KeyLength:  int(answer.GetKeyLength()),
	}

	return dbInfo
}

func runQueries(ctx context.Context, cs map[string]*grpc.ClientConn, queries [][]byte) [][]byte {
	if len(cs) != len(queries) {
		log.Fatal("queries and server addresses length mismatch")
	}

	subCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan []byte, len(cs))
	j := 0
	for _, conn := range cs {
		wg.Add(1)
		go func(j int, conn *grpc.ClientConn) {
			resCh <- query(subCtx, conn, queries[j])
			wg.Done()
		}(j, conn)
		j++
	}
	wg.Wait()
	close(resCh)

	// combinate anwsers of all the servers
	q := make([][]byte, 0)
	for v := range resCh {
		q = append(q, v)
	}

	return q
}

func query(ctx context.Context, conn *grpc.ClientConn, query []byte) []byte {
	c := proto.NewVPIRClient(conn)
	q := &proto.QueryRequest{Query: query}
	var opts []grpc.CallOption
	opts = append(opts, grpc.UseCompressor(gzip.Name))
	answer, err := c.Query(ctx, q, opts...)
	if err != nil {
		log.Fatalf("could not query %s: %v",
			conn.Target(), err)
	}
	log.Printf("sent query to %s", conn.Target())

	return answer.GetAnswer()
}

func equalDBInfo(info []*database.Info) bool {
	for i := range info {
		if info[0].NumRows != info[i].NumRows ||
			info[0].NumColumns != info[i].NumColumns ||
			info[0].BlockSize != info[i].BlockSize ||
			info[0].IDLength != info[i].IDLength ||
			info[0].KeyLength != info[i].KeyLength {
			return false
		}
	}

	return true
}
