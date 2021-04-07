package main

import (
	"fmt"
	"testing"

	dpfgo "github.com/dimakogan/dpf-go/dpf"
	"github.com/si-co/vpir-code/lib/dpf"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestDPFBandwidth(t *testing.T) {
	logN := uint64(28)
	index := uint64(1)
	blockSize := 64

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

func sizeVpirDPFKey(a, b dpf.DPFkey) int {
	// ServerIdx byte
	out := 2
	// Bytes     []byte
	out += len(a.Bytes) + len(b.Bytes)
	// FinalCW   []field.Element
	out += 16*len(a.FinalCW) + 16*len(b.FinalCW)

	return out
}
