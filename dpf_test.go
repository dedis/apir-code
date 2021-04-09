package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	dpfgo "github.com/dimakogan/dpf-go/dpf"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestDPFBandwidth(t *testing.T) {
	logN := uint64(28)
	index := uint64(1)
	blockSize := 128

	// classical PIR
	key0, key1 := dpfgo.Gen(index, logN)
	lenGoDPF := len(key0) + len(key1)
	fmt.Println("Bandwidth godpf:", lenGoDPF)

	// VPIR
	alpha := field.Element{}
	if _, err := alpha.SetRandom(utils.RandomPRG()); err != nil {
		require.NoError(t, err)
	}

	a := make([]field.Element, blockSize+1)
	a[0] = field.One()
	for i := 1; i < len(a); i++ {
		a[i].Mul(&a[i-1], &alpha)
	}
	key00, key11 := dpf.Gen(index, a, logN)
	lenDPF := sizeVpirDPFKey(key00, key11)
	fmt.Println("Bandwidth dpf:", lenDPF)
	fmt.Println("Factor:", float32(lenDPF/lenGoDPF))
}

func TestDPFCPUTime(t *testing.T) {
	logN := uint64(28)
	index := uint64(1)
	blockSize := 128

	// initialize monitor
	m := monitor.NewMonitor()

	// classical PIR
	m.Reset()
	key0, key1 := dpfgo.Gen(index, logN)
	lenGoDPF := len(key0) + len(key1)
	fmt.Println("Classical DPF Gen CPU time", m.RecordAndReset(), "bytes:", lenGoDPF)

	// VPIR
	alpha := field.Element{}
	if _, err := alpha.SetRandom(utils.RandomPRG()); err != nil {
		require.NoError(t, err)
	}

	a := make([]field.Element, blockSize+1)
	a[0] = field.One()
	for i := 1; i < len(a); i++ {
		a[i].Mul(&a[i-1], &alpha)
	}
	m.Reset()
	key00, key11 := dpf.Gen(index, a, logN)
	fmt.Println("Field DPF Gen CPU time", m.RecordAndReset(), "bytes:", sizeVpirDPFKey(key00, key11))
	queries := []dpf.DPFkey{key00, key11}

	m.Reset()
	// encode all the queries in bytes
	out := make([][]byte, len(queries))
	for i, q := range queries {
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		if err := enc.Encode(q); err != nil {
			panic(err)
		}
		out[i] = buf.Bytes()
	}
	fmt.Println("Encoding time:", m.RecordAndReset())
}

func sizeVpirDPFKey(a, b dpf.DPFkey) int {
	// ServerIdx byte
	out := 2
	// Bytes     []byte
	out += len(a.Bytes) + len(b.Bytes)
	//fmt.Println("Bytes length:", len(a.Bytes)+len(b.Bytes))
	// FinalCW   []field.Element
	out += 16*len(a.FinalCW) + 16*len(b.FinalCW)

	return out
}
