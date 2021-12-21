package manager

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
)

// NewManager returns a new initialized manager
func NewManager(config utils.Config, opts []grpc.CallOption) Manager {
	return Manager{
		config: config,
		opts:   opts,
	}
}

// Manager is used to initialize an actor that can manager servers
type Manager struct {
	config utils.Config
	opts   []grpc.CallOption
}

// Connect connects to the server and returns an Actor that can query the
// servers.
func (m *Manager) Connect() (Actor, error) {
	servers := make(map[string]server)

	// load servers certificates
	creds, err := utils.LoadServersCertificates()
	if err != nil {
		return Actor{}, xerrors.Errorf("could not load servers certificates: %v", err)
	}

	for _, addr := range m.config.Addresses {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds),
			grpc.WithBlock())
		if err != nil {
			return Actor{}, xerrors.Errorf("did not connect to %s: %v", addr, err)
		}

		servers[addr] = server{conn: conn, opts: m.opts}
	}

	return Actor{
		servers: servers,
		opts:    m.opts,
	}, nil
}

// Actor allows to perform operations on the servers.
type Actor struct {
	servers map[string]server
	opts    []grpc.CallOption
}

// GetKey perform a simple query that return a key from an email
func (a *Actor) GetKey(id string, dbInfo database.Info, client *client.PIR) (string, error) {
	t := time.Now()

	// compute hash key for id
	hashKey := database.HashToIndex(id, dbInfo.NumRows*dbInfo.NumColumns)
	log.Printf("id: %s, hashKey: %d", id, hashKey)

	// query given hash key
	in := make([]byte, 4)
	binary.BigEndian.PutUint32(in, uint32(hashKey))
	queries, err := client.QueryBytes(in, len(a.servers))
	if err != nil {
		return "", xerrors.Errorf("error when executing query: %v", err)
	}
	log.Printf("done with queries computation")

	// send queries to servers
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan []byte, len(a.servers))
	j := 0

	for _, srv := range a.servers {
		wg.Add(1)
		go func(j int, srv server) {
			resCh <- srv.query(ctx, queries[j])
			wg.Done()
		}(j, srv)
		j++
	}
	wg.Wait()
	close(resCh)

	// combinate answers of all the servers
	answers := make([][]byte, 0)
	for v := range resCh {
		answers = append(answers, v)
	}

	// reconstruct block
	resultField, err := client.ReconstructBytes(answers)
	if err != nil {
		return "", xerrors.Errorf("error during reconstruction: %v", err)
	}
	log.Printf("done with block reconstruction")

	result := resultField.([]byte)
	result = database.UnPadBlock(result)

	// get a key from the block with the id of the search
	retrievedKey, err := pgp.RecoverKeyFromBlock(result, id)
	if err != nil {
		return "", xerrors.Errorf("error retrieving key from the block: %v", err)
	}
	log.Printf("PGP key retrieved from block")

	armored, err := pgp.ArmorKey(retrievedKey)
	if err != nil {
		return "", xerrors.Errorf("error armor-encoding the key: %v", err)
	}

	fmt.Println(armored)

	elapsedTime := time.Since(t)

	fmt.Printf("Wall-clock time to retrieve the key: %v\n", elapsedTime)

	return armored, nil
}

// GetDBInfos returns infos about the servers dbs.
func (a *Actor) GetDBInfos() ([]database.Info, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan database.Info, len(a.servers))

	for _, srv := range a.servers {
		wg.Add(1)
		go func(srv server) {
			defer wg.Done()

			info := srv.getDBInfo(ctx)
			resCh <- info
		}(srv)
	}

	wg.Wait()
	close(resCh)

	dbInfo := make([]database.Info, 0, len(resCh))

	for info := range resCh {
		dbInfo = append(dbInfo, info)
	}

	// check if db info are all equal before returning
	for i := range dbInfo {
		if dbInfo[0].NumRows != dbInfo[i].NumRows ||
			dbInfo[0].NumColumns != dbInfo[i].NumColumns ||
			dbInfo[0].BlockSize != dbInfo[i].BlockSize {

			return nil, xerrors.Errorf("db not equal: %v", dbInfo)
		}
	}

	log.Printf("databaseInfo: %#v", dbInfo[0])

	return dbInfo, nil
}

// RunQueries dispatch queries in parallel to all servers. It then combines the
// answers.
func (a *Actor) RunQueries(queries [][]byte) [][]byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	wg := sync.WaitGroup{}
	resCh := make(chan []byte, len(a.servers))
	j := 0
	for _, srv := range a.servers {
		wg.Add(1)
		go func(j int, srv server) {
			resCh <- srv.query(ctx, queries[j])
			wg.Done()
		}(j, srv)
		j++
	}
	wg.Wait()
	close(resCh)

	// combinate answers of all the servers
	q := make([][]byte, 0)
	for v := range resCh {
		q = append(q, v)
	}

	return q
}

// server represents a remote server
type server struct {
	conn *grpc.ClientConn
	opts []grpc.CallOption
}

// query performs a query on the server
func (s server) query(ctx context.Context, query []byte) []byte {
	c := proto.NewVPIRClient(s.conn)
	q := &proto.QueryRequest{Query: query}

	answer, err := c.Query(ctx, q, s.opts...)
	if err != nil {
		log.Fatalf("could not query %s: %v",
			s.conn.Target(), err)
	}

	log.Printf("sent query to %s", s.conn.Target())
	log.Printf("query size in bytes %d", len(query))

	return answer.GetAnswer()
}

// getDBInfo returns DB info about the server
func (s server) getDBInfo(ctx context.Context) database.Info {
	c := proto.NewVPIRClient(s.conn)
	q := &proto.DatabaseInfoRequest{}

	answer, err := c.DatabaseInfo(ctx, q, s.opts...)
	if err != nil {
		log.Fatalf("could not send database info request to %s: %v",
			s.conn.Target(), err)
	}

	log.Printf("sent databaseInfo request to %s", s.conn.Target())

	dbInfo := database.Info{
		NumRows:    int(answer.GetNumRows()),
		NumColumns: int(answer.GetNumColumns()),
		BlockSize:  int(answer.GetBlockLength()),
		PIRType:    answer.GetPirType(),
		Merkle:     &database.Merkle{Root: answer.GetRoot(), ProofLen: int(answer.GetProofLen())},
	}

	return dbInfo
}
