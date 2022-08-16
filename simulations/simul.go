package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"time"

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
	NumServers     []int
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

	// create results directory if not presenc
	folderPath := "results"
	if _, err := os.Stat(folderPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write mem profile to file")
	indivConfigFile := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// CPU profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
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
		nRows := s.NumRows
		numBlocks := dl

		// matrix db
		if nRows != 1 {
			utils.IncreaseToNextSquare(&numBlocks)
			nRows = int(math.Sqrt(float64(numBlocks)))
		}

		// setup db, this is the same for DPF or IT
		dbPRG := utils.RandomPRG()
		dbRing := new(database.Ring)
		dbElliptic := new(database.Elliptic)
		dbLWE := new(database.LWE)
		switch s.Primitive[:3] {
		case "cmp":
			if s.Primitive == "cmp-pir" {
				log.Printf("Generating lattice db of size %d\n", dbLen)
				dbRing = database.CreateRandomRingDB(dbPRG, dbLen, true)
			} else if s.Primitive == "cmp-vpir" {
				log.Printf("Generating elliptic db of size %d\n", dbLen)
				dbElliptic = database.CreateRandomEllipticWithDigest(dbPRG, dbLen, group.P256, true)
			} else if s.Primitive == "cmp-vpir-lwe" {
				log.Printf("Generating LWE db of size %d\n", dbLen)
				dbLWE = database.CreateRandomBinaryLWEWithLength(dbPRG, dbLen)
			}
		}

		// GC after DB creation
		runtime.GC()
		time.Sleep(3)

		// run experiment
		var results []*Chunk
		switch s.Primitive {
		case "cmp-pir":
			log.Printf("db info: %#v", dbRing.Info)
			results = pirLattice(dbRing, s.Repetitions)
		case "cmp-vpir":
			log.Printf("db info: %#v", dbElliptic.Info)
			results = pirElliptic(dbElliptic, s.Repetitions)
		case "cmp-vpir-lwe":
			log.Printf("db info: %#v", dbLWE.Info)
			results = pirLWE(dbLWE, s.Repetitions)
		case "preprocessing":
			log.Printf("Merkle preprocessing evaluation for dbLen %d bits\n", dbLen)
			results = RandomMerkleDB(dbPRG, dbLen, nRows, blockLen, s.Repetitions)
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

	// mem profiling
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
	log.Println("simulation terminated successfully")
}

func pirLattice(db *database.Ring, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	c := client.NewLattice(&db.Info)
	s := server.NewLattice(db)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index := rand.Intn(db.NumRows * db.NumColumns)
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

func pirLWE(db *database.LWE, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	p := utils.ParamsWithDatabaseSize(db.Info.NumRows, db.Info.NumColumns)
	c := client.NewLWE(utils.RandomPRG(), &db.Info, p)
	s := server.NewLWE(db)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index := rand.Intn(db.NumRows * db.NumColumns)
		results[j].CPU[0] = initBlock(1)
		results[j].Bandwidth[0] = initBlock(1)

		m.Reset()
		query, err := c.QueryBytes(index)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Query = m.RecordAndReset()
		results[j].Bandwidth[0].Query += float64(len(query))

		// get server's answer
		answer, err := s.AnswerBytes(query)
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

func pirElliptic(db *database.Elliptic, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	prg := utils.RandomPRG()
	c := client.NewDH(prg, &db.Info)
	s := server.NewDH(db)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index := rand.Intn(db.NumRows * db.NumColumns)
		results[j].CPU[0] = initBlock(1)
		results[j].Bandwidth[0] = initBlock(1)

		m.Reset()
		query, err := c.QueryBytes(index)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Query = m.RecordAndReset()
		results[j].Bandwidth[0].Query += float64(len(query))

		// get server's answer
		answer, err := s.AnswerBytes(query)
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

func infoSizeByte(q *query.Info) int {
	totalLen := 0
	// Count the bytes of Info
	// q.Target and q.Targets are uint8 and []uint8,
	// respectively
	totalLen += len(q.Targets) + 1 // q.Target
	// The size of int is platform dependent
	totalLen += bits.UintSize / 8 //q.FromStart
	totalLen += bits.UintSize / 8 // q.FromEnd
	// And is bool
	totalLen += 1

	return totalLen
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
	return s.Primitive == "cmp-pir" ||
		s.Primitive == "cmp-vpir" ||
		s.Primitive == "cmp-vpir-lwe" ||
		s.Primitive == "preprocessing"
}
