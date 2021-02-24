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
		if s.BlockLength == constants.SingleBitBlockLength {
			db = database.CreateRandomSingleBitDB(dbPRG, dbLen, nRows)
		} else {
			db = database.CreateRandomMultiBitDB(dbPRG, dbLen, nRows, blockLen)
		}

		// run experiment
		var results []*Chunk
		log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits", s.Primitive, dbLen)
		if s.Primitive == "vpir-it" {
			results = retrieveIT(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
		} else if s.Primitive == "vpir-dpf" {
			results = retrieveDPF(db, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
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

func retrieveIT(db *database.DB, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	cl := client.NewIT(prg, &db.Info)
	servers := makeITServers(db)

	var numBlocksToRetrieve int
	if db.BlockSize == constants.SingleBitBlockLength {
		numBlocksToRetrieve = numBitsToRetrieve
	} else {
		numBlocksToRetrieve = numBitsToRetrieve / (db.BlockSize * elemBitSize)
	}

	return retrieveBlocks(cl, servers, db.NumRows*db.NumColumns, numBlocksToRetrieve, nRepeat)
}

func retrieveDPF(db *database.DB, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	cl := client.NewDPF(prg, &db.Info)
	servers := makeDPFServers(db)

	var numBlocksToRetrieve int
	if db.BlockSize == constants.SingleBitBlockLength {
		numBlocksToRetrieve = numBitsToRetrieve
	} else {
		numBlocksToRetrieve = numBitsToRetrieve / (db.BlockSize * elemBitSize)
	}

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
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Query = m.RecordAndReset()
			for r := range queries {
				results[j].Bandwidth[i].Query += float64(len(queries[r]))
			}

			// get servers answers
			answers := make([][]byte, len(ss))
			for k := range ss {
				answers[k], err = ss[k].AnswerBytes(queries[k])
				if err != nil {
					log.Fatal(err)
				}
				results[j].CPU[i].Answers[k] = m.RecordAndReset()
				results[j].Bandwidth[i].Answers[k] = float64(len(answers[k]))
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

func (s *Simulation) validSimulation() bool {
	return s.Primitive == "vpir-it" || s.Primitive == "vpir-dpf" || s.Primitive == "pir"
}

//func retrieveBlocksDPF(db *database.DB, numBlocks int, nRepeat int) []*Chunk {
//	prg := utils.RandomPRG()
//	c := client.NewDPF(prg, &db.Info)
//	s0 := server.NewDPF(db)
//	s1 := server.NewDPF(db)
//
//	// create main monitor for CPU time
//	m := monitor.NewMonitor()
//
//	// run the experiment nRepeat times
//	results := make([]*Chunk, nRepeat)
//
//	for j := 0; j < nRepeat; j++ {
//		log.Printf("start repetition %d out of %d", j+1, nRepeat)
//		results[j] = &Chunk{
//			CPU: make([]*Block, numBlocks),
//		}
//
//		totalTimer := monitor.NewMonitor()
//		for i := 0; i < numBlocks; i++ {
//			results[j].CPU[i] = new(Block)
//
//			m.Reset()
//			queries := c.Query(i, 2)
//			results[j].CPU[i].Query = m.RecordAndReset()
//
//			a0 := s0.Answer(queries[0])
//			results[j].CPU[i].Answer0 = m.RecordAndReset()
//
//			a1 := s1.Answer(queries[1])
//			results[j].CPU[i].Answer1 = m.RecordAndReset()
//
//			answers := [][]field.Element{a0, a1}
//
//			m.Reset()
//			_, err := c.Reconstruct(answers)
//			results[j].CPU[i].Reconstruct = m.RecordAndReset()
//			if err != nil {
//				panic(err)
//			}
//
//		}
//		results[j].TotalCPU = totalTimer.Record()
//	}
//
//	return results
//}

//func retrieveBlocksIT(db *database.DB, elemBitSize int, numBitsToRetrieve int, nRepeat int) []*Chunk {
//	prg := utils.RandomPRG()
//	c := client.NewIT(prg, &db.Info)
//	s0 := server.NewIT(db)
//	s1 := server.NewIT(db)
//
//	var numBlocksToRetrieve int
//	if db.BlockSize == constants.SingleBitBlockLength {
//		numBlocksToRetrieve = numBitsToRetrieve
//	} else {
//		numBlocksToRetrieve = numBitsToRetrieve / (db.BlockSize * elemBitSize)
//	}
//
//	// seed non-cryptographic randomness
//	rand.Seed(time.Now().UnixNano())
//
//	// create main monitor for CPU time
//	m := monitor.NewMonitor()
//
//	// run the experiment nRepeat times
//	results := make([]*Chunk, nRepeat)
//
//	var startIndex int
//	for j := 0; j < nRepeat; j++ {
//		log.Printf("start repetition %d out of %d", j+1, nRepeat)
//		results[j] = &Chunk{
//			CPU: make([]*Block, numBlocksToRetrieve),
//		}
//		// pick a random block index to start the retrieval
//		startIndex = rand.Intn(db.NumRows*db.NumColumns - numBlocksToRetrieve)
//		totalTimer := monitor.NewMonitor()
//		for i := 0; i < numBlocksToRetrieve; i++ {
//			results[j].CPU[i] = new(Block)
//
//			m.Reset()
//			queries := c.Query(startIndex+i, 2)
//			results[j].CPU[i].Query = m.RecordAndReset()
//
//			a0 := s0.Answer(queries[0])
//			results[j].CPU[i].Answer0 = m.RecordAndReset()
//
//			a1 := s1.Answer(queries[1])
//			results[j].CPU[i].Answer1 = m.RecordAndReset()
//
//			answers := [][]field.Element{a0, a1}
//
//			m.Reset()
//			_, err := c.Reconstruct(answers)
//			results[j].CPU[i].Reconstruct = m.RecordAndReset()
//			if err != nil {
//				log.Fatal(err)
//			}
//		}
//		results[j].TotalCPU = totalTimer.Record()
//	}
//
//	return results
//}
