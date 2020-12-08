package main

import (
	"fmt"
	"testing"

	"github.com/ncw/gmp"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

func TestITRetrieval(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiDatabase()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	c := client.NewITClient(xof)
	s0 := server.NewITServer(db)
	s1 := server.NewITServer(db)
	s2 := server.NewITServer(db)
	m := monitor.NewMonitor()
	for i := 0; i < 136; i++ {
		m.Reset()
		queries := c.Query(i, 3)
		fmt.Printf("Query: %.3fms\t", m.RecordAndReset())

		a0 := s0.Answer(queries[0])
		fmt.Printf("Answer 1: %.3fms\t", m.RecordAndReset())

		a1 := s1.Answer(queries[1])
		fmt.Printf("Answer 2: %.3fms\t", m.RecordAndReset())

		a2 := s2.Answer(queries[2])
		fmt.Printf("Answer 3: %.3fms\t", m.RecordAndReset())

		answers := make([]*gmp.Int, 3)
		answers[0] = a0
		answers[1] = a1
		answers[2] = a2

		m.Reset()
		x, err := c.Reconstruct(answers)
		fmt.Printf("Reconstruct: %.3fms\n", m.RecordAndReset())
		if err != nil {
			t.Error(err)
			panic(err)
		}
		result += x.String()
	}
	b, err := utils.BitStringToBytes(result)
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

	fmt.Printf("Total time: %.1fms\n", totalTimer.Record())
}
