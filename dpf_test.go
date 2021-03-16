package main

import (
	"fmt"
	"testing"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
)

func TestEvalFull(t *testing.T) {
	dbLen := 8000000000 // 1GB
	blockLen := testBlockLength
	elemSize := 128
	numBlocks := dbLen / (elemSize * blockLen)
	nRows := 1

	xofDB := utils.RandomPRG()
	xof := utils.RandomPRG()
	db := database.CreateRandomMultiBitDB(xofDB, dbLen, nRows, blockLen)

	c := client.NewDPF(xof, &db.Info)
	s0 := server.NewDPF(db)

	totalTimer := monitor.NewMonitor()
	time := float64(0)
	for i := 0; i < numBlocks; i++ {
		fssKeys := c.Query(i, 2)

		totalTimer.Reset()
		_ = s0.Answer(fssKeys[0])
		time += totalTimer.RecordAndReset()

	}

	totalTime := time
	fmt.Printf("Total CPU time per %d queries: %fms\n", numBlocks, totalTime)
	fmt.Printf("Throughput: %f GB/s\n", float64(numBlocks)/(totalTime*0.001))
}
