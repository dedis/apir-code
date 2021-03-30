package dpf

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/utils"
	"github.com/stretchr/testify/require"
)

func TestEvalFull(t *testing.T) {
	toSec := 0.001
	databaseBytes := (1 << 30)

	// Number of field elements in each block
	bytesPerFieldElement := 15
	blockLen := 16
	nRows := 1

	db, err := database.CreateRandomMultiBitDB(utils.RandomPRG(), databaseBytes*8, nRows, blockLen)
	require.NoError(t, err)

	alpha, err := new(field.Element).SetRandom(utils.RandomPRG())
	if err != nil {
		panic(err)
	}

	beta := make([]field.Element, blockLen+1)
	beta[0] = field.One()
	for i := 1; i < len(beta); i++ {
		beta[i].Mul(&beta[i-1], alpha)
	}

	fmt.Printf("Num columns: %d\n", db.NumColumns)
	fmt.Printf("Num rows: %d\n", nRows)
	time := float64(0)
	key0, _ := Gen(7, beta, uint64(bits.Len(uint(db.NumColumns)-1)))
	q := make([]field.Element, db.NumColumns*(db.BlockSize+1))

	dpfTimer := monitor.NewMonitor()
	dpfTimer.Reset()
	EvalFullFlatten(key0, uint64(bits.Len(uint(db.NumColumns)-1)), len(beta), q)
	totalTime := dpfTimer.RecordAndReset()

	fmt.Printf("Total CPU time per %d queries: %fms\n", 1, totalTime)
	fmt.Printf("Throughput EvalFull: %f GB/s\n", float64(databaseBytes)/(totalTime*toSec)/float64(1<<30))

	// AES
	prfkeyL := []byte{36, 156, 50, 234, 92, 230, 49, 9, 174, 170, 205, 160, 98, 236, 29, 243}
	keyL := make([]uint32, 11*4)
	expandKeyAsm(&prfkeyL[0], &keyL[0])
	dst := new(block)
	src := new(block)

	aesBytesPerBlock := 16
	aesBlocks := databaseBytes / (aesBytesPerBlock)
	aesTimer := monitor.NewMonitor()
	time = 0
	aesTimer.Reset()
	for i := 0; i < aesBlocks; i++ {
		aes128MMO(&keyL[0], &dst[0], &src[0])
	}
	time += aesTimer.RecordAndReset()

	totalTime = time
	fmt.Printf("Total CPU time per %d AES blocks: %fms\n", aesBlocks, totalTime)
	fmt.Printf("Throughput AES: %f GB/s\n", float64(databaseBytes)/(totalTime*toSec)/float64(1<<30))

	// Field operations
	prg := utils.RandomPRG()
	a, err := new(field.Element).SetRandom(prg)
	if err != nil {
		panic(err)
	}
	b, err := new(field.Element).SetRandom(prg)
	if err != nil {
		panic(err)
	}

	fieldElements := databaseBytes / bytesPerFieldElement
	fieldTimer := monitor.NewMonitor()
	fieldTimer.Reset()
	for i := 0; i < fieldElements; i++ {
		a.Mul(a, b)
	}
	totalTime = fieldTimer.RecordAndReset()
	fmt.Printf("Total CPU time per %d field ops: %fms\n", fieldElements, totalTime)
	fmt.Printf("Throughput field ops: %f GB/s\n", float64(databaseBytes)/(totalTime*toSec)/float64(1<<30))
}
