package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
)

type Simulation struct {
	DBLengthsBits  []float64
	Repetitions    int
	Name           string
	Primitive      string
	NumRows        int
	BlockLength    int
	ElementBitSize int
}

func main() {
	configFile := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// make sure cfg file is specified
	if *configFile == "" {
		panic("simulation's config file not provided")
	}
	log.Printf("config file %s", *configFile)

	// load simulation's config file
	s := new(Simulation)
	_, err := toml.DecodeFile(*configFile, s)
	if err != nil {
		panic(err)
	}
	log.Printf("running simulation %#v\n", s)

	// check simulation
	if !s.validSimulation() {
		panic("invalid simulation")
	}

	// initialize experiment
	experiment := &Experiment{Results: make(map[int][]*DBResult, 0)}

	// range over all the DB lengths specified in the general simulation config
	for _, dl := range s.DBLengthsBits {
		// compute database data
		dbLen := int(dl)
		blockLen := s.BlockLength
		elemBitSize := s.ElementBitSize
		nRows := s.NumRows
		var nCols int
		// vector case
		if nRows == 1 {
			if s.BlockLength == constants.SingleBitBlockLength {
				nCols = dbLen
			} else {
				nCols = dbLen / (elemBitSize * blockLen * nRows)
			}
		} else {
			var numBlocks int
			if s.BlockLength == constants.SingleBitBlockLength {
				numBlocks = dbLen
			} else {
				numBlocks = dbLen / (elemBitSize * blockLen)
			}
			utils.IncreaseToNextSquare(&numBlocks)
			// for really small db
			if numBlocks == 0 {
				numBlocks = 1
			}
			nCols = int(math.Sqrt(float64(numBlocks)))
			nRows = nCols

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
		log.Printf("retrieving blocks from DB with dbLen = %d bits", dbLen)
		results := retrieveBlocksIT(db, nCols, s.Repetitions)
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

	log.Println("simulation terminated succesfully")
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

func retrieveBlocksIT(db *database.DB, numBlocks int, nRepeat int) []*DBResult {
	prg := utils.RandomPRG()
	c := client.NewIT(prg, &db.Info)
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)

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

func (s *Simulation) validSimulation() bool {
	return s.Primitive == "vpir" || s.Primitive == "pir"
}
