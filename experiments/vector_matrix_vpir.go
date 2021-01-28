package main

import (
	"fmt"

	"github.com/si-co/vpir-code/lib/field_gcm"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
	"golang.org/x/crypto/blake2b"
)

func main() {
	// repeat the experiments nRepeat times
	nRepeat := 10

	// setup 1kb random db
	db := database.CreateAsciiMatrixOneKb()

	// setup local clients and servers
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := true
	vpir := true
	c := client.NewITSingleGF(xof, rebalanced, vpir)
	s0 := server.NewITSingleGF(rebalanced, db)
	s1 := server.NewITSingleGF(rebalanced, db)

	// times taken in ms
	fmt.Println("query,answer0,answer1,answer2,reconstruct")

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	for j := 0; j < nRepeat; j++ {
		totalTimer := monitor.NewMonitor()
		for i := 0; i < 8191; i++ {
			m.Reset()
			queries := c.Query(i, 3)
			fmt.Printf("%.3f,", m.RecordAndReset())

			a0 := s0.Answer(queries[0])
			fmt.Printf("%.3f,", m.RecordAndReset())

			a1 := s1.Answer(queries[1])
			fmt.Printf("%.3f,", m.RecordAndReset())

			answers := [][]field_gcm.Element{a0, a1}

			m.Reset()
			_, err := c.Reconstruct(answers)
			fmt.Printf("%.3f,", m.RecordAndReset())
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("%.3f\n", totalTimer.Record())
	}
}
