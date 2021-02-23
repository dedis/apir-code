package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"path"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
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
	indivConfigFile := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// make sure cfg file is specified
	if *indivConfigFile == "" {
		panic("simulation's config file not provided")
	}
	log.Printf("config file %s", *indivConfigFile)

	// load simulation's config files
	var err error
	genConfig := new(generalParam)
	_, err = toml.DecodeFile(generalConfigFile, genConfig)
	if err != nil {
		log.Fatal(err)
	}
	indConfig := new(individualParam)
	_, err = toml.DecodeFile(*indivConfigFile, indConfig)
	if err != nil {
		log.Fatal(err)
	}
	s := &Simulation{generalParam: *genConfig, individualParam: *indConfig}

	log.Printf("running simulation %#v\n", s)

	// check simulation
	if !s.validSimulation() {
		panic("invalid simulation")
	}

	// initialize experiment
	experiment := &Experiment{Results: make(map[int][]*DBResult, 0)}

	// range over all the DB lengths specified in the general simulation config
	for _, dl := range s.DBBitLengths {
		// compute database data
		dbLen := dl
		blockLen := s.BlockLength
		elemBitSize := s.ElementBitSize
		nRows := s.NumRows

		var numBlocks int
		var nCols int
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
		nCols = numBlocks / nRows

		// setup db, this is the same for DPF or IT
		dbPRG := utils.RandomPRG()
		db := new(database.DB)
		if s.BlockLength == constants.SingleBitBlockLength {
			db = database.CreateRandomSingleBitDB(dbPRG, dbLen, nRows)
		} else {
			db = database.CreateRandomMultiBitDB(dbPRG, dbLen, nRows, blockLen)
		}

		// run experiment
		var results []*DBResult
		log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits", s.Primitive, dbLen)
		if s.Primitive == "vpir-it" {
			results = retrieveBlocksIT(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		} else if s.Primitive == "vpir-dpf" {
			results = retrieveBlocksDPF(db, nCols, s.Repetitions)
		} else {
			panic("not yet implemented")
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

func retrieveBlocksDPF(db *database.DB, numBlocks int, nRepeat int) []*DBResult {
	prg := utils.RandomPRG()
	c := client.NewDPF(prg, &db.Info)
	s0 := server.NewDPF(db)
	s1 := server.NewDPF(db)

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := make([]*DBResult, nRepeat)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = &DBResult{
			Results: make([]*BlockResult, numBlocks),
		}

		totalTimer := monitor.NewMonitor()
		for i := 0; i < numBlocks; i++ {
			results[j].Results[i] = new(BlockResult)

			m.Reset()
			queries := c.Query(i, 2)
			results[j].Results[i].Query = m.RecordAndReset()

			a0 := s0.Answer(queries[0])
			results[j].Results[i].Answer0 = m.RecordAndReset()

			a1 := s1.Answer(queries[1])
			results[j].Results[i].Answer1 = m.RecordAndReset()

			answers := [][]field.Element{a0, a1}

			m.Reset()
			_, err := c.Reconstruct(answers)
			results[j].Results[i].Reconstruct = m.RecordAndReset()
			if err != nil {
				panic(err)
			}

		}
		results[j].Total = totalTimer.Record()
	}

	return results
}

func retrieveBlocksIT(db *database.DB, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*DBResult {
	prg := utils.RandomPRG()
	c := client.NewIT(prg, &db.Info)
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)

	var numBlocksToRetrieve int
	if db.BlockSize == constants.SingleBitBlockLength {
		numBlocksToRetrieve = numBitsToRetrieve
	} else {
		numBlocksToRetrieve = numBitsToRetrieve / (db.BlockSize * elemBitSize)
	}

	// seed non-cryptographic randomness
	rand.Seed(time.Now().UnixNano())

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := make([]*DBResult, nRepeat)

	var startIndex int
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = &DBResult{
			Results: make([]*BlockResult, numBlocksToRetrieve),
		}
		// pick a random block index to start the retrieval
		startIndex = rand.Intn(db.NumRows*db.NumColumns - numBlocksToRetrieve)
		totalTimer := monitor.NewMonitor()
		for i := 0; i < numBlocksToRetrieve; i++ {
			results[j].Results[i] = new(BlockResult)

			m.Reset()
			queries := c.Query(startIndex+i, 2)
			results[j].Results[i].Query = m.RecordAndReset()

			a0 := s0.Answer(queries[0])
			results[j].Results[i].Answer0 = m.RecordAndReset()

			a1 := s1.Answer(queries[1])
			results[j].Results[i].Answer1 = m.RecordAndReset()

			answers := [][]field.Element{a0, a1}

			m.Reset()
			_, err := c.Reconstruct(answers)
			results[j].Results[i].Reconstruct = m.RecordAndReset()
			if err != nil {
				log.Fatal(err)
			}
		}
		results[j].Total = totalTimer.Record()
	}

	return results
}

func (s *Simulation) validSimulation() bool {
	return s.Primitive == "vpir-it" || s.Primitive == "vpir-dpf" || s.Primitive == "pir"
}
