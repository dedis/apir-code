package main

import (
	"encoding/json"
	"fmt"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
)

const (
	oneMB = 1048576 * 8
	oneKB = 1024 * 8
)

type BlockResult struct {
	Query       float64
	Answer0     float64
	Answer1     float64
	Reconstruct float64
}

type DBResult struct {
	BlockResults []*BlockResult
	Total        float64
}

type ExperimentResults struct {
	Results []*DBResult
}

func main() {
	// repeat the experiments nRepeat times
	nRepeat := 10

	// database data
	dbLen := oneMB
	blockLen := constants.BlockLength
	elemBitSize := field.Bytes * 8
	nRows := 1
	nCols := dbLen / (elemBitSize * blockLen * nRows)

	// create db
	db := database.CreateRandomMultiBitDB(utils.RandomPRG(), dbLen, nRows, blockLen)

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
