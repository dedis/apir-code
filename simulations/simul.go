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
	Name           string
	Primitive      string
	DBLengthsBits  []float64
	NumRows        int
	BlockLength    int
	ElementBitSize int
	Repetitions    int
}

func main() {
	configFile := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// make sure cfg file is specified
	if *configFile == "" {
		panic("simulation's config file not provided")
	}

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
	experiment := &Experiment{Results: make([]*DBResult, 0)}

	for _, dl := range s.DBLengthsBits {
		// compute database data
		dbLen := int(dl * 1000000.0)
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
			nCols = int(math.Sqrt(float64(numBlocks)))
			nRows = nCols
		}

		// setup db
		dbPRG := utils.RandomPRG()
		db := new(database.DB)
		if s.BlockLength == constants.SingleBitBlockLength {
			db = database.CreateRandomSingleBitDB(dbPRG, dbLen, nRows)
		} else {
			db = database.CreateRandomMultiBitDB(dbPRG, dbLen, nRows, blockLen)
		}

		// run experiment
		results := retrieveBlocks(db, dbLen, nCols, s.Repetitions)

		// append result to general experiment
		experiment.Results = append(experiment.Results, results...)
	}

	// print results
	res, err := json.Marshal(experiment)
	if err != nil {
		panic(err)
	}
	if err = ioutil.WriteFile(path.Join("results", s.Name, ".json"), res, 0644); err != nil {
		panic(err)
	}

	log.Println("simulation terminated succesfully")
}

func retrieveBlocks(db *database.DB, dbLen int, numBlocks int, nRepeat int) []*DBResult {
	log.Printf("retrieving blocks from DB with dbLen = %d bits", dbLen)

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
			Results:      make([]*BlockResult, numBlocks),
			DBLengthBits: dbLen,
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
