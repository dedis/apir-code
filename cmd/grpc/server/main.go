package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"

	"github.com/si-co/vpir-code/lib/codec"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/pgp"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"

	"github.com/si-co/vpir-code/lib/proto"
	"github.com/si-co/vpir-code/lib/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	configEnvKey = "VPIR_CONFIG"
	dataEnvKey   = "VPIR_SKS_ROOT"

	defaultConfigFile = "config.toml"
	defaultSksPath    = "data"
)

func main() {
	// flags
	sid := flag.Int("id", -1, "Server ID")
	experiment := flag.Bool("experiment", false, "run setting for experiments")
	filesNumber := flag.Int("files", 1, "number of key files to use in db creation")
	cores := flag.Int("cores", -1, "number of cores to use")
	vpirScheme := flag.String("scheme", "", "scheme to use: it, dpf, pir-it or pir-dpf")
	logFile := flag.String("log", "", "write log to file instead of stdout/stderr")
	prof := flag.Bool("prof", false, "Write CPU prof file")
	mprof := flag.Bool("mprof", false, "Write memory prof file")

	flag.Parse()

	// start profiling
	if *prof {
		utils.StartProfiling(fmt.Sprintf("server-%v.prof", *sid))
		defer utils.StopProfiling()
	}

	if *mprof {
		fn := fmt.Sprintf("server-%v-mem.mprof", *sid)
		defer func() {
			f, err := os.Create(fn)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Writing memory profile")
			pprof.WriteHeapProfile(f)
			f.Close()
		}()
	}

	// set logs
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("[Server %v] ", *sid))
	if len(*logFile) > 0 {
		f, err := os.Create(*logFile)
		if err != nil {
			log.Fatal("Could not open file: ", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// configs
	configPath := os.Getenv(configEnvKey)
	if configPath == "" {
		configPath = defaultConfigFile
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("could not load the server config file: %v", err)
	}
	addr := config.Addresses[*sid]

	// load the db
	var db *database.DB
	var dbBytes *database.Bytes
	switch *vpirScheme {
	case "it":
		db, err = loadPgpDB(*filesNumber, true)
		if err != nil {
			log.Fatalf("impossible to load real keys db: %v", err)
		}
		log.Printf("db size in GiB: %f", db.SizeGiB())
	case "dpf":
		db, err = loadPgpDB(*filesNumber, false)
		if err != nil {
			log.Fatalf("impossible to construct real keys db: %v", err)
		}
		log.Printf("db size in GiB: %f", db.SizeGiB())
	case "pir-it":
		dbBytes, err = loadPgpBytes(*filesNumber, true)
		if err != nil {
			log.Fatalf("impossible to construct real keys bytes db: %v", err)
		}
		log.Printf("db size in GiB: %f", dbBytes.SizeGiB())
	case "pir-dpf":
		dbBytes, err = loadPgpBytes(*filesNumber, false)
		if err != nil {
			log.Fatalf("impossible to construct real keys bytes db: %v", err)
		}
		log.Printf("db size in GiB: %f", dbBytes.SizeGiB())
	case "merkle-it":
		dbBytes, err = loadPgpMerkle(*filesNumber, true)
		if err != nil {
			log.Fatalf("impossible to construct real keys bytes db: %v", err)
		}
		log.Printf("db size in GiB: %f", dbBytes.SizeGiB())
	case "merkle-dpf":
		dbBytes, err = loadPgpMerkle(*filesNumber, false)
		if err != nil {
			log.Fatalf("impossible to construct real keys bytes db: %v", err)
		}
		log.Printf("db size in GiB: %f", dbBytes.SizeGiB())
	default:
		log.Fatal("unknown scheme")
	}

	// GC after db creation
	runtime.GC()

	// run server with TLS
	cfg := &tls.Config{
		Certificates: []tls.Certificate{utils.ServerCertificates[*sid]},
		ClientAuth:   tls.NoClientCert,
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer(
		// grpc.MaxRecvMsgSize(1024*1024*1024),
		// grpc.MaxSendMsgSize(1024*1024*1024),
		grpc.Creds(credentials.NewTLS(cfg)),
		grpc.CustomCodec(&codec.Codec{}),
	)

	// select correct server
	var s server.Server
	switch *vpirScheme {
	case "it":
		if *cores != -1 && *experiment {
			s = server.NewIT(db, *cores)
		} else {
			s = server.NewIT(db)
		}
	case "dpf":
		if *cores != -1 && *experiment {
			s = server.NewDPF(db, *cores)
		} else {
			s = server.NewDPF(db)
		}
	case "pir-it", "merkle-it":
		if *cores != -1 && *experiment {
			s = server.NewPIR(dbBytes, *cores)
		} else {
			s = server.NewPIR(dbBytes)
		}
	case "pir-dpf", "merkle-dpf":
		if *cores != -1 && *experiment {
			s = server.NewPIRdpf(dbBytes, *cores)
		} else {
			s = server.NewPIRdpf(dbBytes)
		}
	default:
		log.Fatal("unknow VPIR type")
	}

	// start server
	proto.RegisterVPIRServer(rpcServer, &vpirServer{
		Server:     s,
		experiment: *experiment,
		cores:      *cores,
	})
	log.Printf("is listening at %s", addr)

	// listen signals from os
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error, 1)

	go func() {
		log.Println("starting grpc server")
		if err := rpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	// start HTTP server for tests
	// TODO: remove this in application
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal("impossible to parse addr for HTTP server")
	}
	h := func(w http.ResponseWriter, _ *http.Request) {
		sigCh <- os.Interrupt
	}
	httpAddr := fmt.Sprintf("%s:%s", host, "8080")
	srv := &http.Server{Addr: httpAddr}
	http.HandleFunc("/", h)
	go func() {
		srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		log.Fatalf("failed to serve: %v", err)
	case <-sigCh:
		rpcServer.GracefulStop()
		lis.Close()
		srv.Shutdown(context.Background())
		log.Println("clean shutdown of server done")
	}
}

// vpirServer is used to implement VPIR Server protocol.
type vpirServer struct {
	proto.UnimplementedVPIRServer
	Server server.Server // both IT and DPF-based server

	// only for experiments
	experiment bool
	cores      int
}

func (s *vpirServer) DatabaseInfo(ctx context.Context, r *proto.DatabaseInfoRequest) (
	*proto.DatabaseInfoResponse, error) {
	log.Print("got databaseInfo request")

	dbInfo := s.Server.DBInfo()
	resp := &proto.DatabaseInfoResponse{
		NumRows:     uint32(dbInfo.NumRows),
		NumColumns:  uint32(dbInfo.NumColumns),
		BlockLength: uint32(dbInfo.BlockSize),
		PirType:     dbInfo.PIRType,
		Root:        dbInfo.Root,
		ProofLen:    uint32(dbInfo.ProofLen),
	}

	return resp, nil
}

// Standard naive one
// func (s *vpirServer) QueryStream(srv proto.VPIR_QueryStreamServer) error {
// 	info := s.Server.DBInfo()
// 	n := info.BlockSize * info.NumColumns * info.NumRows

// 	data := make([]byte, 0, n)

// 	for i := 0; i < s.Server.DBInfo().NumColumns; i++ {
// 		req, err := srv.Recv()
// 		if err != nil {
// 			return xerrors.Errorf("failed to read request: %v", err)
// 		}

// 		data = append(data, req.Query...)
// 	}

// 	a, err := s.Server.AnswerBytes(data)
// 	if err != nil {
// 		return err
// 	}

// 	answerLen := len(a)
// 	log.Printf("answer size in bytes: %d", answerLen)
// 	if s.experiment {
// 		log.Printf("stats,%d,%d", s.cores, answerLen)
// 	}

// 	srv.SendAndClose(&proto.QueryResponse{Answer: a})

// 	return nil
// }

func newWorkerPool(size int) *workerPool {
	return &workerPool{
		size:    size,
		working: 0,

		workerAvailable: make(chan struct{}, 1),
	}
}

type workerPool struct {
	sync.Mutex

	size    int
	working int

	// used by worker to signal that there are done
	workerAvailable chan struct{}
}

func (w *workerPool) addAndWait(f func()) {
	w.Lock()

	if w.size == w.working {
		w.Unlock()
		<-w.workerAvailable
		w.Lock()
	}

	go func() {
		w.working++
		w.Unlock()

		f()

		w.Lock()
		w.working--

		select {
		case w.workerAvailable <- struct{}{}:
		default:
		}

		w.Unlock()
	}()
}

type homework struct {
	i int
	q *proto.QueryRequest
}

func newWorkers(inputs <-chan homework, outputs chan<- []field.Element, blockSize int, s *server.IT) *workers {
	return &workers{
		inputs:  inputs,
		outputs: outputs,

		loadPerWorker: 10,
		blockSize:     blockSize,

		server: s,
		stop:   make(chan struct{}),

		finished: new(sync.WaitGroup),
	}
}

type workers struct {
	inputs  <-chan homework
	outputs chan<- []field.Element

	loadPerWorker int
	blockSize     int

	server *server.IT

	stop chan struct{}

	finished *sync.WaitGroup
}

func (w workers) start(n int) {
	for i := 0; i < n; i++ {
		w.finished.Add(1)
		go w.startWorker()
	}

	go func() {
		w.finished.Wait()
		close(w.outputs)
	}()
}

func (w workers) startWorker() {
	defer w.finished.Done()

	for {
		res := []field.Element{}

		for i := 0; i < w.loadPerWorker; i++ {
			var homework homework

			select {
			case homework = <-w.inputs:
			case <-w.stop:
				// be sure there isn't anything left in the input
				select {
				case homework = <-w.inputs:
				default:
					if len(res) != 0 {
						w.outputs <- res
					}
					return
				}
			}

			elements := field.NewElemSliceFromBytes(homework.q.Query)
			r := w.server.ComputeMessageAndTagNew(homework.i*w.blockSize, (homework.i+1)*w.blockSize, elements, w.blockSize)

			if len(res) != 0 {
				for i, a := range r {
					res[i].Add(&res[i], &a)
				}
			} else {
				res = r
			}
		}

		w.outputs <- res
	}
}

func (w workers) done() {
	close(w.stop)
}

func (s *vpirServer) QueryStream(srv proto.VPIR_QueryStreamServer) error {
	res := []field.Element{}

	inQueue := make(chan homework, 50)
	outQueue := make(chan []field.Element, 50)

	info := s.Server.DBInfo()

	workers := newWorkers(inQueue, outQueue, info.BlockSize, s.Server.(*server.IT))
	workers.start(10)

	outDone := sync.WaitGroup{}
	outDone.Add(1)
	go func() {
		defer outDone.Done()

		for r := range outQueue {
			if len(res) == 0 {
				res = r
				continue
			}

			for i, a := range r {
				res[i].Add(&res[i], &a)
			}
		}
	}()

	for i := 0; i < info.NumColumns; i++ {
		req, err := srv.Recv()
		if err != nil {
			return xerrors.Errorf("failed to read request: %v", err)
		}

		inQueue <- homework{
			i: i,
			q: req,
		}
	}

	workers.done()
	outDone.Wait()

	answerLen := len(res)
	log.Printf("answer size in bytes: %d", answerLen)
	if s.experiment {
		log.Printf("stats,%d,%d", s.cores, answerLen)
	}

	buf := make([]byte, len(res)*8*2)
	for k := 0; k < len(res); k++ {
		binary.LittleEndian.PutUint64(buf[k*8*2:k*8*2+8], res[k][0])
		binary.LittleEndian.PutUint64(buf[k*8*2+8:k*8*2+8+8], res[k][1])
	}

	srv.SendAndClose(&proto.QueryResponse{Answer: buf})

	return nil
}

// func (s *vpirServer) QueryStream(srv proto.VPIR_QueryStreamServer) error {
// 	info := s.Server.DBInfo()

// 	var err error

// 	requests := make([][]byte, info.NumColumns)

// 	for i := 0; i < s.Server.DBInfo().NumColumns; i++ {
// 		req, err := srv.Recv()
// 		if err != nil {
// 			return xerrors.Errorf("failed to read request: %v", err)
// 		}

// 		requests[i] = req.Query
// 	}

// 	elemGetter := field.NewElemSliceGetter(
// 		field.NewBytesChunks(requests, len(requests[0])),
// 	)

// 	a, err := s.Server.(*server.IT).AnswerBytesNew(elemGetter)
// 	if err != nil {
// 		return err
// 	}
// 	answerLen := len(a)
// 	log.Printf("answer size in bytes: %d", answerLen)
// 	if s.experiment {
// 		log.Printf("stats,%d,%d", s.cores, answerLen)
// 	}

// 	srv.SendAndClose(&proto.QueryResponse{Answer: a})

// 	return nil
// }

func (s *vpirServer) Query(ctx context.Context, qr *proto.QueryRequest) (
	*proto.QueryResponse, error) {
	log.Printf("got query request, len(Query): %d", len(qr.Query))

	a, err := s.Server.AnswerBytes(qr.GetQuery())
	if err != nil {
		return nil, err
	}
	answerLen := len(a)
	log.Printf("answer size in bytes: %d", answerLen)
	if s.experiment {
		log.Printf("stats,%d,%d", s.cores, answerLen)
	}

	return &proto.QueryResponse{Answer: a}, nil
}

func loadPgpDB(filesNumber int, rebalanced bool) (*database.DB, error) {
	log.Println("Starting to read in the DB data")

	// take only filesNumber files
	// files := getSksFiles(filesNumber)

	// db, err := database.GenerateRealKeyDB(files, constants.ChunkBytesLength, rebalanced)
	db, err := database.LoadMMapDB(os.Getenv(dataEnvKey))
	if err != nil {
		return nil, err
	}
	// log.Println("DB loaded with files", files)

	return db, nil
}

func loadPgpBytes(filesNumber int, rebalanced bool) (*database.Bytes, error) {
	log.Println("Starting to read in the DB data")

	// take only filesNumber files
	files := getSksFiles(filesNumber)

	db, err := database.GenerateRealKeyBytes(files, rebalanced)
	if err != nil {
		return nil, err
	}
	log.Println("Bytes loaded with files", files)

	return db, nil
}

func loadPgpMerkle(filesNumber int, rebalanced bool) (*database.Bytes, error) {
	log.Println("Starting to read in the DB data")

	// take only filesNumber files
	files := getSksFiles(filesNumber)

	db, err := database.GenerateRealKeyMerkle(files, rebalanced)
	if err != nil {
		return nil, err
	}
	log.Println("Bytes loaded with files", files)

	return db, nil
}

func getSksFiles(filesNumber int) []string {
	sksDir := os.Getenv(dataEnvKey)
	if sksDir == "" {
		sksDir = filepath.Join(defaultSksPath, pgp.SksParsedFolder)
	}

	files, err := pgp.GetAllFiles(sksDir)
	if err != nil {
		log.Fatalf("impossible to get sks files: %v", err)
	}
	// take only filesNumber files
	return files[:filesNumber]
}
