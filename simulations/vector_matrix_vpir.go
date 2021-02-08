package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
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
	cfg := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// make sure cfg file is specified
	if *cfg == "" {
		panic("simulation's config file not provided")
	}

	// load simulation's config file
	s := new(Simulation)
	_, err := toml.DecodeFile(configFile, s)
	if err != nil {
		panic(err)
	}

	// check simulation
	if s.Primitive != "vpir" && s.Primitive != "pir" {
		panic("unsupported primitive")
	}

	// database data
	dbLen := s.DBLengthBits
	blockLen := s.BlockLength
	elemBitSize := s.ElementBitSize
	nRows := s.NumRows
	// TODO: fix nCols for single-bit schemes
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	// setup db
	dbPRG := utils.RandomPRG()
	if s.BlockLength == constants.SingleBlockLength {
		db := database.CreateRandomSingleBitDB(dbPRG, dbLen, nRows)
	} else {
		db := database.CreateRandomMultiBitDB(dbPRG, dbLen, nRows, blockLen)
	}

	// run experiment
	retrieveBlocks(db, nCols, nRepeat)
}

func retrieveBlocks(db *database.DB, numBlocks int, nRepeat int) {
	prg := utils.RandomPRG()
	c := client.NewIT(prg, &db.Info)
	s0 := server.NewIT(db)
	s1 := server.NewIT(db)

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := &ExperimentResults{Results: make([]*DBResult, nRepeat)}

	for j := 0; j < nRepeat; j++ {
		results.Results[j] = &DBResult{BlockResults: make([]*BlockResult, numBlocks)}
		totalTimer := monitor.NewMonitor()
		for i := 0; i < numBlocks; i++ {
			results.Results[j].BlockResults[i] = new(BlockResult)

			m.Reset()
			queries := c.Query(i, 2)
			//fmt.Printf("%.3f,", m.RecordAndReset())
			results.Results[j].BlockResults[i].Query = m.RecordAndReset()

			a0 := s0.Answer(queries[0])
			//fmt.Printf("%.3f,", m.RecordAndReset())
			results.Results[j].BlockResults[i].Answer0 = m.RecordAndReset()

			a1 := s1.Answer(queries[1])
			//fmt.Printf("%.3f,", m.RecordAndReset())
			results.Results[j].BlockResults[i].Answer1 = m.RecordAndReset()

			answers := [][]field.Element{a0, a1}

			m.Reset()
			_, err := c.Reconstruct(answers)
			//fmt.Printf("%.3f,", m.RecordAndReset())
			results.Results[j].BlockResults[i].Reconstruct = m.RecordAndReset()
			if err != nil {
				panic(err)
			}

		}
		results.Results[j].Total = totalTimer.Record()
		//fmt.Printf("%.3f\n", totalTimer.Record())

	}

	res, err := json.Marshal(results)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(res))

}
