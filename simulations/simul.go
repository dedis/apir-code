package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"

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
	DBLengthBits   int
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

	// check simulation
	if !s.validSimulation() {
		panic("invalid simulation")
	}

	// compute database data
	dbLen := s.DBLengthBits
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
	retrieveBlocks(db, nCols, s.Repetitions)
}

func retrieveBlocks(db *database.DB, numBlocks int, nRepeat int) {
	prg := utils.RandomPRG()
	c := client.NewIT(prg, &db.Info)
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := &simul.Experiment{Results: make([]*simul.DBResult, nRepeat)}

	for j := 0; j < nRepeat; j++ {
		results.Results[j] = &simul.DBResult{Results: make([]*simul.BlockResult, numBlocks)}
		totalTimer := monitor.NewMonitor()
		for i := 0; i < numBlocks; i++ {
			results.Results[j].Results[i] = new(simul.BlockResult)

			m.Reset()
			queries := c.Query(i, 2)
			results.Results[j].Results[i].Query = m.RecordAndReset()

			a0 := s0.Answer(queries[0])
			results.Results[j].Results[i].Answer0 = m.RecordAndReset()

			a1 := s1.Answer(queries[1])
			results.Results[j].Results[i].Answer1 = m.RecordAndReset()

			answers := [][]field.Element{a0, a1}

			m.Reset()
			_, err := c.Reconstruct(answers)
			results.Results[j].Results[i].Reconstruct = m.RecordAndReset()
			if err != nil {
				panic(err)
			}

		}
		results.Results[j].Total = totalTimer.Record()
	}

	// print results
	res, err := json.Marshal(results)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(res))

}

func (s *Simulation) validSimulation() bool {
	return s.Primitive == "vpir" || s.Primitive == "pir"
}
