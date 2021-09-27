package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
)

const generalConfigFile = "simul.toml"

type generalParam struct {
	DBBitLengths   []int
	BitsToRetrieve int
	Repetitions    int
}

type individualParam struct {
	Name           string
	Primitive      string
	NumRows        int
	BlockLength    int
	ElementBitSize int
	InputSizes     []int // FSS input sizes in bytes
}

type Simulation struct {
	generalParam
	individualParam
}

func main() {
	// seed non-cryptographic randomness
	rand.Seed(time.Now().UnixNano())

	// tracing
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	indivConfigFile := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// CPU profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// make sure cfg file is specified
	if *indivConfigFile == "" {
		panic("simulation's config file not provided")
	}
	log.Printf("config file %s", *indivConfigFile)

	// load simulation's config files
	s, err := loadSimulationConfigs(generalConfigFile, *indivConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	// check simulation
	if !s.validSimulation() {
		log.Fatal("invalid simulation")
	}

	log.Printf("running simulation %#v\n", s)
	// initialize experiment
	experiment := &Experiment{Results: make(map[int][]*Chunk, 0)}

	// TODO (Simone): what if we range on DB sizes for point and single and on input bit size
	// for complex?
	// range over all the DB lengths specified in the general simulation config
dbSizesLoop:
	for _, dl := range s.DBBitLengths {
		// compute database data
		dbLen := dl
		blockLen := s.BlockLength
		elemBitSize := s.ElementBitSize
		nRows := s.NumRows

		var numBlocks int
		// Find the total number of blocks in the db
		numBlocks = dbLen / (elemBitSize * blockLen)
		// matrix db
		if nRows != 1 {
			utils.IncreaseToNextSquare(&numBlocks)
			nRows = int(math.Sqrt(float64(numBlocks)))
		}

		// setup db, this is the same for DPF or IT
		dbPRG := utils.RandomPRG()
		db := new(database.DB)
		dbBytes := new(database.Bytes)
		dbRing := new(database.Ring)
		dbElliptic := new(database.Elliptic)
		switch s.Primitive[:3] {
		case "pir":
			if s.Primitive[len(s.Primitive)-6:] == "merkle" {
				dbBytes = database.CreateRandomMerkle(dbPRG, dbLen, nRows, blockLen)
			} else {
				dbBytes = database.CreateRandomBytes(dbPRG, dbLen, nRows, blockLen)
			}
		case "fss":
			// TODO: update config or db creation to match dbLen params, or vice versa
			// TODO (Simone): I would fix the number of identifiers
			// here to e.g. 100k, instead of using the numBlocks
			// variable
			numIdenfitiers := 100000
			//db, err = database.CreateRandomKeysDB(dbPRG, numBlocks)
			db, err = database.CreateRandomKeysDB(dbPRG, numIdenfitiers)
			if err != nil {
				panic(err)
			}
		case "cmp":
			if s.Primitive == "cmp-pir" {
				log.Printf("Generating lattice db of size %d\n", dbLen)
				dbRing = database.CreateRandomRingDB(dbPRG, dbLen, true)
			} else if s.Primitive == "cmp-vpir" {
				log.Printf("Generating elliptic db of size %d\n", dbLen)
				dbElliptic = database.CreateRandomEllipticWithDigest(dbPRG, dbLen, group.P256, true)
			}
		}

		// GC after DB creation
		runtime.GC()
		time.Sleep(3)

		// run experiment
		var results []*Chunk
		switch s.Primitive {
		case "pir-classic", "pir-merkle":
			log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits",
				s.Primitive, dbLen)
			// TODO (Simone): avoid printing BlockLengths, too verbose
			//log.Printf("db info: %#v", dbBytes.Info)
			blockSize := dbBytes.BlockSize - dbBytes.ProofLen // ProofLen = 0 for PIR
			results = pirIT(dbBytes, blockSize, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		case "fss-pir":
			// TODO (Simone): avoid printing BlockLengths, too verbose
			//log.Printf("db info: %#v", db.Info)
			// In FSS, we iterate over input sizes instead of db sizes
			for _, inputSize := range s.InputSizes {
				results = fssPIR(db, inputSize, s.Repetitions)
				experiment.Results[inputSize] = results
			}
			// Skip the rest of the loop
			break dbSizesLoop
		case "fss-vpir":
			// TODO (Simone): avoid printing BlockLengths, too verbose
			//log.Printf("db info: %#v", db.Info)
			for _, inputSize := range s.InputSizes {
				results = fssVPIR(db, inputSize, s.Repetitions)
				experiment.Results[inputSize] = results
			}
			// Skip the rest of the loop
			break dbSizesLoop
		case "cmp-pir":
			log.Printf("db info: %#v", dbRing.Info)
			results = pirLattice(dbRing, s.Repetitions)
		case "cmp-vpir":
			log.Printf("db info: %#v", dbElliptic.Info)
			results = pirElliptic(dbElliptic, s.Repetitions)
		default:
			log.Fatal("unknown primitive type")
		}
		experiment.Results[dbLen] = results

		// GC at the end of the iteration
		runtime.GC()
	}

	// print results
	res, err := json.Marshal(experiment)
	if err != nil {
		panic(err)
	}
	fileName := s.Name + ".json"
	if err = ioutil.WriteFile(path.Join("results", fileName), res, 0644); err != nil {
		panic(err)
	}

	log.Println("simulation terminated successfully")
}

func fssVPIR(db *database.DB, inputSize int, nRepeat int) []*Chunk {
	c := client.NewFSS(utils.RandomPRG(), &db.Info)
	ss := []*server.FSS{server.NewFSS(db, 0), server.NewFSS(db, 1)}

	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	// TODO (Simone): is this the best way to evaluate the FSS-based
	// approach? Or should have some form of determinism here?
	stringToSearch := utils.Ranstring(inputSize)

	in := utils.ByteToBits([]byte(stringToSearch))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: inputSize},
		Input: in,
	}
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(1)
		results[j].CPU[0] = initBlock(len(ss))
		results[j].Bandwidth[0] = initBlock(len(ss))

		m.Reset()
		queries := c.Query(q, 2)
		results[j].CPU[0].Query = m.RecordAndReset()
		for r := range queries {
			results[j].Bandwidth[0].Query += fssQueryByteLength(queries[r])
		}

		// get servers answers
		answers := make([][]uint32, len(ss))
		for k := range ss {
			m.Reset()
			answers[k] = ss[k].Answer(queries[k])
			results[j].CPU[0].Answers[k] = m.RecordAndReset()
			results[j].Bandwidth[0].Answers[k] = fieldVectorByteLength(answers[k])
		}

		m.Reset()
		_, err := c.Reconstruct(answers)
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		if err != nil {
			log.Fatal(err)
		}
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func fssPIR(db *database.DB, inputSize int, nRepeat int) []*Chunk {
	c := client.NewPIRfss(utils.RandomPRG(), &db.Info)
	ss := []*server.PIRfss{server.NewPIRfss(db, 0), server.NewPIRfss(db, 1)}

	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	stringToSearch := utils.Ranstring(inputSize)

	in := utils.ByteToBits([]byte(stringToSearch))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: inputSize},
		Input: in,
	}
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(1)
		results[j].CPU[0] = initBlock(len(ss))
		results[j].Bandwidth[0] = initBlock(len(ss))

		m.Reset()
		queries := c.Query(q, 2)
		results[j].CPU[0].Query = m.RecordAndReset()
		for r := range queries {
			results[j].Bandwidth[0].Query += fssQueryByteLength(queries[r])
		}

		// get servers answers
		answers := make([][]uint32, len(ss))
		for k := range ss {
			m.Reset()
			answers[k] = ss[k].Answer(queries[k])
			results[j].CPU[0].Answers[k] = m.RecordAndReset()
			results[j].Bandwidth[0].Answers[k] = fieldVectorByteLength(answers[k])
		}

		m.Reset()
		_, err := c.Reconstruct(answers)
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		if err != nil {
			log.Fatal(err)
		}
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func pirIT(db *database.Bytes, blockSize, elemBitSize, numBitsToRetrieve, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	c := client.NewPIR(prg, &db.Info)
	ss := makePIRServers(db)
	numTotalBlocks := db.NumRows * db.NumColumns
	numRetrieveBlocks := bitsToBlocks(blockSize, elemBitSize, numBitsToRetrieve)

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	var startIndex int
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrieveBlocks)

		// pick a random block index to start the retrieval
		startIndex = rand.Intn(numTotalBlocks - numRetrieveBlocks)
		for i := 0; i < numRetrieveBlocks; i++ {
			results[j].CPU[i] = initBlock(len(ss))
			results[j].Bandwidth[i] = initBlock(len(ss))

			m.Reset()
			queries := c.Query(startIndex+i, 2)
			results[j].CPU[i].Query = m.RecordAndReset()
			for r := range queries {
				results[j].Bandwidth[i].Query += float64(len(queries[r]))
			}

			// get servers answers
			answers := make([][]byte, len(ss))
			for k := range ss {
				m.Reset()
				answers[k] = ss[k].Answer(queries[k])
				results[j].CPU[i].Answers[k] = m.RecordAndReset()
				results[j].Bandwidth[i].Answers[k] = float64(len(answers[k]))
			}

			m.Reset()
			_, err := c.Reconstruct(answers)
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			results[j].Bandwidth[i].Reconstruct = 0
			if err != nil {
				log.Fatal(err)
			}
		}

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func pirLattice(db *database.Ring, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	c := client.NewLattice(&db.Info)
	s := server.NewLattice(db)

	var index int
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index = rand.Intn(db.NumRows * db.NumColumns)
		for i := 0; i < numRetrievedBlocks; i++ {
			results[j].CPU[i] = initBlock(1)
			results[j].Bandwidth[i] = initBlock(1)

			m.Reset()
			query, err := c.QueryBytes(index + i)
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Query = m.RecordAndReset()
			results[j].Bandwidth[i].Query = float64(len(query))

			answer, err := s.AnswerBytes(query)
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Answers[0] = m.RecordAndReset()
			results[j].Bandwidth[i].Answers[0] = float64(len(answer))

			_, err = c.ReconstructBytes(answer)
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			results[j].Bandwidth[i].Reconstruct = 0
		}

		// GC after each repetition
		runtime.GC()
		time.Sleep(2)
	}
	return results
}

func pirElliptic(db *database.Elliptic, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	prg := utils.RandomPRG()
	c := client.NewDH(prg, &db.Info)
	s := server.NewDH(db)

	var index int
	var err error
	var query, answer []byte
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index = rand.Intn(db.NumRows * db.NumColumns)
		results[j].CPU[0] = initBlock(1)
		results[j].Bandwidth[0] = initBlock(1)

		m.Reset()
		query, err = c.QueryBytes(index)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Query = m.RecordAndReset()
		results[j].Bandwidth[0].Query += float64(len(query))

		// get server's answer
		answer, err = s.AnswerBytes(query)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Answers[0] = m.RecordAndReset()
		results[j].Bandwidth[0].Answers[0] = float64(len(answer))

		_, err = c.ReconstructBytes(answer)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()
		time.Sleep(2)
	}

	return results
}

// Converts number of bits to retrieve into the number of db blocks
func bitsToBlocks(blockSize, elemSize, numBits int) int {
	return int(math.Ceil(float64(numBits) / float64(blockSize*elemSize)))
}

func makePIRServers(db *database.Bytes) []*server.PIR {
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)
	return []*server.PIR{s0, s1}
}

func fssQueryByteLength(q *query.AuthFSS) float64 {
	totalLen := 0

	// Count the bytes of FssKey
	totalLen += len(q.FssKey.SInit)
	totalLen += 1 // q.FssKey.TInit
	totalLen += len(q.FssKey.FinalCW) * field.Bytes
	for i := range q.FssKey.CW {
		totalLen += len(q.FssKey.CW[i])
	}

	// Count the bytes of AdditionalInformationFSS
	// q.Target and q.Targets are uint8 and []uint8,
	// respectively
	totalLen += len(q.Targets) + 1 // q.Target
	// The size of int is platform dependent
	totalLen += int(unsafe.Sizeof(q.FromStart))
	totalLen += int(unsafe.Sizeof(q.FromEnd))
	// And is bool
	totalLen += 1

	return float64(totalLen)
}

func fieldVectorByteLength(vec []uint32) float64 {
	return float64(len(vec) * field.Bytes)
}

func initChunk(numRetrieveBlocks int) *Chunk {
	return &Chunk{
		CPU:       make([]*Block, numRetrieveBlocks),
		Bandwidth: make([]*Block, numRetrieveBlocks),
	}
}

func initBlock(numAnswers int) *Block {
	return &Block{
		Query:       0,
		Answers:     make([]float64, numAnswers),
		Reconstruct: 0,
	}
}

func loadSimulationConfigs(genFile, indFile string) (*Simulation, error) {
	var err error
	genConfig := new(generalParam)
	_, err = toml.DecodeFile(genFile, genConfig)
	if err != nil {
		return nil, err
	}
	indConfig := new(individualParam)
	_, err = toml.DecodeFile(indFile, indConfig)
	if err != nil {
		return nil, err
	}
	return &Simulation{generalParam: *genConfig, individualParam: *indConfig}, nil
}

func (s *Simulation) validSimulation() bool {
	return s.Primitive == "pir-classic" ||
		s.Primitive == "pir-merkle" ||
		s.Primitive == "fss-pir" ||
		s.Primitive == "fss-vpir" ||
		s.Primitive == "cmp-pir" ||
		s.Primitive == "cmp-vpir"
}
