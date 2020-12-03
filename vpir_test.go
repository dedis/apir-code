package main

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"golang.org/x/crypto/blake2b"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/timer"
)

func TestAlgorithm(t *testing.T) {
	//start := time.Now()
	totalTimer := timer.NewMonitor()
	db := database.CreateAsciiDatabase()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	c := client.NewClient(xof)
	s0 := server.CreateServer(db)
	s1 := server.CreateServer(db)
	s2 := server.CreateServer(db)
	for i := 0; i < 136; i++ {
		startQuery := time.Now()
		queries, st := c.Query(i, 3)
		elapsedQuery := time.Since(startQuery)
		fmt.Printf("Query took %s \n", elapsedQuery)

		startAnswerFirst := time.Now()
		a0 := s0.Answer(queries[0])
		elapsedAnswerFirst := time.Since(startAnswerFirst)
		fmt.Printf("Answer first took %s \n", elapsedAnswerFirst)

		startAnswerSecond := time.Now()
		a1 := s1.Answer(queries[1])
		elapsedAnswerSecond := time.Since(startAnswerSecond)
		fmt.Printf("Answer second took %s \n", elapsedAnswerSecond)

		startAnswerThird := time.Now()
		a2 := s2.Answer(queries[2])
		elapsedAnswerThird := time.Since(startAnswerThird)
		fmt.Printf("Answer third took %s \n", elapsedAnswerThird)

		answers := make([]*big.Int, 3)
		answers[0] = a0
		answers[1] = a1
		answers[2] = a2

		startReconstruct := time.Now()
		x, err := c.Reconstruct(answers, st)
		elapsedReconstruct := time.Since(startReconstruct)
		fmt.Printf("Reconstruct took %s \n", elapsedReconstruct)
		if err != nil {
			t.Error(err)
			panic(err)
		}
		result += x.String()
	}
	b, err := client.BitStringToBytes(result)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	output := string(b)
	fmt.Println(output)

	const expected = "Playing with VPIR"
	if expected != output {
		t.Errorf("Expected '%v' but got '%v'", expected, output)
	}

	//elapsed := time.Since(start)
	elapsed := totalTimer.Record()
	fmt.Printf("Took %.1fms", elapsed)
}
