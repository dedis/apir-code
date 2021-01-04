package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"

	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/server"
)

func TestMatrixOneKbByte(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiMatrixOneKbByte()
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := true
	c := client.NewITSingleByte(xof, rebalanced)
	s0 := server.NewITSingleByte(rebalanced, db)
	s1 := server.NewITSingleByte(rebalanced, db)
	s2 := server.NewITSingleByte(rebalanced, db)
	for i := 0; i < 8191; i++ {
		queries := c.Query(i, 3)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])
		a2 := s2.Answer(queries[2])

		answers := [][]byte{a0, a1, a2}

		_, err := c.Reconstruct(answers)
		require.NoError(t, err)
	}
	fmt.Printf("Total time MatrixOneKbByte: %.1fms\n", totalTimer.Record())
}

func TestMatrixOneKbGF(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiMatrixOneKb()
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := true
	vpir := true
	c := client.NewITSingleGF(xof, rebalanced, vpir)
	s0 := server.NewITSingleGF(rebalanced, db)
	s1 := server.NewITSingleGF(rebalanced, db)
	s2 := server.NewITSingleGF(rebalanced, db)
	for i := 0; i < 8191; i++ {
		queries := c.Query(i, 3)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])
		a2 := s2.Answer(queries[2])

		answers := [][]field.Element{a0, a1, a2}

		_, err := c.Reconstruct(answers)
		require.NoError(t, err)
	}
	fmt.Printf("Total time MatrixOneKbGF: %.1fms\n", totalTimer.Record())
}

func TestMatrixGF(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiMatrixGF()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := true
	vpir := true
	c := client.NewITSingleGF(xof, rebalanced, vpir)
	s0 := server.NewITSingleGF(rebalanced, db)
	s1 := server.NewITSingleGF(rebalanced, db)
	s2 := server.NewITSingleGF(rebalanced, db)
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

		answers := [][]field.Element{a0, a1, a2}

		m.Reset()
		x, err := c.Reconstruct(answers)
		require.NoError(t, err)
		fmt.Printf("Reconstruct: %.3fms\n", m.RecordAndReset())
		if x.String() == "00" {
			result += "0"
		} else {
			result += "1"
		}
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

func TestVectorGF(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiVectorGF()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := false
	vpir := true
	c := client.NewITSingleGF(xof, rebalanced, vpir)
	s0 := server.NewITSingleGF(rebalanced, db)
	s1 := server.NewITSingleGF(rebalanced, db)
	s2 := server.NewITSingleGF(rebalanced, db)
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

		answers := [][]field.Element{a0, a1, a2}

		m.Reset()
		x, err := c.Reconstruct(answers)
		require.NoError(t, err)
		fmt.Printf("Reconstruct: %.3fms\n", m.RecordAndReset())
		if x.String() == "00" {
			result += "0"
		} else {
			result += "1"
		}
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

	fmt.Printf("Total time VectorGF: %.1fms\n", totalTimer.Record())
}

func TestVectorByte(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiVectorByte()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	rebalanced := false
	c := client.NewITSingleByte(xof, rebalanced)
	s0 := server.NewITSingleByte(rebalanced, db)
	s1 := server.NewITSingleByte(rebalanced, db)
	s2 := server.NewITSingleByte(rebalanced, db)
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

		answers := [][]byte{a0, a1, a2}

		m.Reset()
		x, err := c.Reconstruct(answers)
		fmt.Println(x)
		require.NoError(t, err)
		fmt.Printf("Reconstruct: %.3fms\n", m.RecordAndReset())
		if x == byte(0) {
			result += "0"
		} else {
			result += "1"
		}
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

	fmt.Printf("Total time VectorByte: %.1fms\n", totalTimer.Record())
}

/**
func TestDPF(t *testing.T) {
	totalTimer := monitor.NewMonitor()
	db := database.CreateAsciiVector()
	result := ""
	xof, err := blake2b.NewXOF(0, []byte("my key"))
	if err != nil {
		panic(err)
	}
	c := client.NewDPF(xof)
	s0 := server.NewDPFServer(db)
	s1 := server.NewDPFServer(db)
	m := monitor.NewMonitor()

	for i := 0; i < 136; i++ {
		m.Reset()
		prfKeys, fssKeys := c.Query(i, 2)
		fmt.Printf("Query: %.3fms\t", m.RecordAndReset())

		a0 := s0.Answer(fssKeys[0], prfKeys, 0)
		fmt.Printf("Answer 1: %.3fms\t", m.RecordAndReset())

		a1 := s1.Answer(fssKeys[1], prfKeys, 1)
		fmt.Printf("Answer 2: %.3fms\t", m.RecordAndReset())

		answers := []*big.Int{a0, a1}

		m.Reset()
		x, err := c.Reconstruct(answers)
		fmt.Printf("Reconstruct: %.3fms\n", m.RecordAndReset())
		if err != nil {
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
**/
