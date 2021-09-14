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

	"github.com/BurntSushi/toml"
	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
)

const generalConfigFile = "simul.toml"

type generalParam struct {
	DBBitLengths   []int
	Repetitions    int
	BitsToRetrieve int
}

type individualParam struct {
	Name           string
	Primitive      string
	NumRows        int
	BlockLength    int
	ElementBitSize int
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

	// range over all the DB lengths specified in the general simulation config
	for _, dl := range s.DBBitLengths {
		// compute database data
		dbLen := dl
		blockLen := s.BlockLength
		elemBitSize := s.ElementBitSize
		nRows := s.NumRows

		var numBlocks int
		// Find the total number of blocks in the db
		numBlocks = dbLen / (elemBitSize * blockLen)
		// for really small db
		if numBlocks == 0 {
			numBlocks = 1
		}

		// rebalanced db
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
		case "vpi":
			db, err = database.CreateRandomMultiBitDB(dbPRG, dbLen, nRows, blockLen)
			if err != nil {
				panic(err)
			}

		case "pir":
			if s.Primitive[len(s.Primitive)-6:] == "merkle" {
				dbBytes = database.CreateRandomMerkle(dbPRG, dbLen, nRows, blockLen)
			} else {
				dbBytes = database.CreateRandomBytes(dbPRG, dbLen, nRows, blockLen)
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
		log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits", s.Primitive, dbLen)
		switch s.Primitive {
		case "vpir-it":
			log.Printf("db info: %#v", db.Info)
			results = vpirIT(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		case "vpir-dpf":
			log.Printf("db info: %#v", db.Info)
			results = vpirDPF(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		case "pir-it", "pir-it-merkle":
			log.Printf("db info: %#v", dbBytes.Info)
			blockSize := dbBytes.BlockSize - dbBytes.ProofLen // ProofLen = 0 for PIR
			results = pirIT(dbBytes, blockSize, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		case "pir-dpf", "pir-dpf-merkle":
			log.Printf("db info: %#v", dbBytes.Info)
			blockSize := dbBytes.BlockSize - dbBytes.ProofLen // ProofLen = 0 for PIR
			results = pirDPF(dbBytes, blockSize, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
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

func vpirIT(db *database.DB, elemBitSize, numBitsToRetrieve, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	c := client.NewIT(prg, &db.Info)
	ss := makeITServers(db)
	numTotalBlocks := db.NumRows * db.NumColumns
	numRetrieveBlocks := bitsToBlocks(db.BlockSize, elemBitSize, numBitsToRetrieve)

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
				results[j].Bandwidth[i].Query += lenBytesFieldSlice(queries[r])
			}

			// get servers answers
			answers := make([][]field.Element, len(ss))
			for k := range ss {
				m.Reset()
				answers[k] = ss[k].Answer(queries[k])
				results[j].CPU[i].Answers[k] = m.RecordAndReset()
				results[j].Bandwidth[i].Answers[k] = lenBytesFieldSlice(answers[k])
			}

			m.Reset()
			_, err := c.Reconstruct(answers)
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			if err != nil {
				log.Fatal(err)
			}
			results[j].Bandwidth[i].Reconstruct = 0
		}

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func vpirDPF(db *database.DB, elemBitSize, numBitsToRetrieve, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	c := client.NewDPF(prg, &db.Info)
	ss := makeDPFServers(db)
	numTotalBlocks := db.NumRows * db.NumColumns
	numRetrieveBlocks := bitsToBlocks(db.BlockSize, elemBitSize, numBitsToRetrieve)

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
			// TODO: len DPF queries
			for r := range queries {
				results[j].Bandwidth[i].Query += lenBytesDPFFieldKey(queries[r])
			}

			// get servers answers
			answers := make([][]field.Element, len(ss))
			for k := range ss {
				m.Reset()
				answers[k] = ss[k].Answer(queries[k])
				results[j].CPU[i].Answers[k] = m.RecordAndReset()
				results[j].Bandwidth[i].Answers[k] = lenBytesFieldSlice(answers[k])
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

func pirIT(db *database.Bytes, blockSize, elemBitSize, numBitsToRetrieve, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	c := client.NewPIR(prg, &db.Info)
	ss := makePIRITServers(db)
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

func pirDPF(db *database.Bytes, blockSize, elemBitSize, numBitsToRetrieve, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	c := client.NewPIRdpf(prg, &db.Info)
	ss := makePIRDPFServers(db)
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
				// key of binary DPF is simply []byte
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
			_, err := c.ReconstructBytes(answers)
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
	return numBits / (blockSize * elemSize)
}

func makeDPFServers(db *database.DB) []*server.DPF {
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)
	return []*server.DPF{s0, s1}
}

func makeITServers(db *database.DB) []*server.IT {
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)
	return []*server.IT{s0, s1}
}

func makePIRITServers(db *database.Bytes) []*server.PIR {
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)
	return []*server.PIR{s0, s1}
}

func makePIRDPFServers(db *database.Bytes) []*server.PIRdpf {
	s0 := server.NewPIRdpf(db)
	s1 := server.NewPIRdpf(db)
	return []*server.PIRdpf{s0, s1}
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

func lenBytesFieldSlice(in []field.Element) float64 {
	return float64(field.Bytes * len(in))
}

func lenBytesDPFFieldKey(in dpf.DPFkey) float64 {
	// ServerIdx byte
	out := float64(1)
	// Bytes     []byte
	out += float64(len(in.Bytes))
	// FinalCW   []field.Element
	out += lenBytesFieldSlice(in.FinalCW)

	return out
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
	return s.Primitive == "vpir-it" ||
		s.Primitive == "vpir-dpf" ||
		s.Primitive == "pir-it" ||
		s.Primitive == "pir-dpf" ||
		s.Primitive == "pir-it-merkle" ||
		s.Primitive == "pir-dpf-merkle" ||
		s.Primitive == "cmp-pir" ||
		s.Primitive == "cmp-vpir"
}
