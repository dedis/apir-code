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
	"runtime/pprof"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
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
		if s.BlockLength == constants.SingleBitBlockLength {
			numBlocks = dbLen
		} else {
			numBlocks = dbLen / (elemBitSize * blockLen)
			// for really small db
			if numBlocks == 0 {
				numBlocks = 1
			}
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
		if s.Primitive[:4] == "vpir" {
			if s.BlockLength == constants.SingleBitBlockLength {
				db = database.CreateRandomSingleBitDB(dbPRG, dbLen, nRows)
			} else {
				db = database.CreateRandomMultiBitDB(dbPRG, dbLen, nRows, blockLen)
			}
		} else if s.Primitive[:3] == "pir" {
			dbBytes = database.CreateRandomMultiBitBytes(dbPRG, dbLen, nRows, blockLen)
		}

		// run experiment
		var results []*Chunk
		log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits", s.Primitive, dbLen)
		if s.Primitive == "vpir-it" {
			results = vpirIT(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		} else if s.Primitive == "vpir-dpf" {
			results = vpirDPF(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		} else if s.Primitive == "pir-it" {
			results = pirIT(dbBytes, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		} else if s.Primitive == "pir-dpf" {
			results = pirDPF(dbBytes, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		} else {
			log.Fatal("unknown primitive type")
		}
		experiment.Results[dbLen] = results
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

func vpirIT(db *database.DB, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	cl := client.NewIT(prg, &db.Info)
	servers := makeITServers(db)
	numBlocksToRetrieve := bitsToBlocks(db.BlockSize, elemBitSize, numBitsToRetrieve)

	return retrieveBlocks(cl, servers, db.NumRows*db.NumColumns, numBlocksToRetrieve, nRepeat)
}

func vpirDPF(db *database.DB, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	cl := client.NewDPF(prg, &db.Info)
	servers := makeDPFServers(db)
	numBlocksToRetrieve := bitsToBlocks(db.BlockSize, elemBitSize, numBitsToRetrieve)

	return retrieveBlocks(cl, servers, db.NumRows*db.NumColumns, numBlocksToRetrieve, nRepeat)
}

func pirIT(db *database.Bytes, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	cl := client.NewPIR(prg, &db.Info)
	servers := makePIRITServers(db)
	numBlocksToRetrieve := bitsToBlocks(db.BlockSize, elemBitSize, numBitsToRetrieve)

	return retrieveBlocks(cl, servers, db.NumRows*db.NumColumns, numBlocksToRetrieve, nRepeat)
}

func pirDPF(db *database.Bytes, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	cl := client.NewPIRdpf(prg, &db.Info)
	servers := makePIRDPFServers(db)
	numBlocksToRetrieve := bitsToBlocks(db.BlockSize, elemBitSize, numBitsToRetrieve)

	return retrieveBlocks(cl, servers, db.NumRows*db.NumColumns, numBlocksToRetrieve, nRepeat)
}

func retrieveBlocks(c client.Client, ss []server.Server, totalBlocks, retrieveBlocks, nRepeat int) []*Chunk {
	// seed non-cryptographic randomness
	rand.Seed(time.Now().UnixNano())

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	var startIndex int
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = &Chunk{
			CPU:       make([]*Block, retrieveBlocks),
			Bandwidth: make([]*Block, retrieveBlocks),
		}
		// pick a random block index to start the retrieval
		startIndex = rand.Intn(totalBlocks - retrieveBlocks)
		for i := 0; i < retrieveBlocks; i++ {
			results[j].CPU[i] = &Block{
				Query:       0,
				Answers:     make([]float64, len(ss)),
				Reconstruct: 0,
			}
			results[j].Bandwidth[i] = &Block{
				Query:       0,
				Answers:     make([]float64, len(ss)),
				Reconstruct: 0,
			}

			m.Reset()
			queries, err := c.QueryBytes(startIndex+i, 2)
			results[j].CPU[i].Query = m.RecordAndReset()
			for r := range queries {
				results[j].Bandwidth[i].Query += float64(len(queries[r]))
			}
			if err != nil {
				log.Fatal(err)
			}

			// get servers answers
			answers := make([][]byte, len(ss))
			for k := range ss {
				answers[k], err = ss[k].AnswerBytes(queries[k])
				results[j].CPU[i].Answers[k] = m.RecordAndReset()
				results[j].Bandwidth[i].Answers[k] = float64(len(answers[k]))
				if err != nil {
					log.Fatal(err)
				}
			}

			_, err = c.ReconstructBytes(answers)
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			results[j].Bandwidth[i].Reconstruct = 0
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return results
}

// Converts number of bits to retrieve into the number of db blocks
func bitsToBlocks(blockSize, elemSize, numBits int) int {
	if blockSize == constants.SingleBitBlockLength {
		return numBits
	}
	return numBits / (blockSize * elemSize)
}

func makeDPFServers(db *database.DB) []server.Server {
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)
	return []server.Server{s0, s1}
}

func makeITServers(db *database.DB) []server.Server {
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)
	return []server.Server{s0, s1}
}

func makePIRITServers(db *database.Bytes) []server.Server {
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)
	return []server.Server{s0, s1}
}

func makePIRDPFServers(db *database.Bytes) []server.Server {
	s0 := server.NewPIRdpf(db)
	s1 := server.NewPIRdpf(db)
	return []server.Server{s0, s1}
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
	return s.Primitive == "vpir-it" || s.Primitive == "vpir-dpf" || s.Primitive == "pir-it" || s.Primitive == "pir-dpf"
}
