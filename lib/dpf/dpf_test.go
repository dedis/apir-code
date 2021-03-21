package dpf

import (
	"fmt"
	"math/bits"
	"testing"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/utils"
)

/*
func BenchmarkEvalFull(b *testing.B) {
	// define db data
	dbLen := 80000000 // 0.01GB
	blockLen := 16
	nRows := 1 // use vector for DPF

	db := database.CreateRandomMultiBitDB(utils.RandomPRG(), dbLen, nRows, blockLen)

	// sample random alpha
	alpha, err := new(field.Element).SetRandom(utils.RandomPRG())
	if err != nil {
		panic(err)
	}

	// compute beta
	beta := make([]field.Element, blockLen+1)
	beta[0] = field.One()
	for i := 1; i < len(beta); i++ {
		beta[i].Mul(&beta[i-1], alpha)
	}

	// create a single key
	key0, _ := Gen(uint64(1), beta, uint64(bits.Len(uint(db.NumColumns)-1)))
	q := make([]field.Element, db.NumColumns*(db.BlockSize+1))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		EvalFullFlatten(key0, uint64(bits.Len(uint(db.NumColumns)-1)), db.BlockSize+1, q)
	}
}
*/

func TestEvalFull(t *testing.T) {
	toSec := 0.001
	databaseBytes := (1 << 30)

  // Number of field elements in each block
  bytesPerFieldElement := 15
  blockLen := 16
	nRows := 1

	db := database.CreateRandomMultiBitDB(utils.RandomPRG(), databaseBytes*8, nRows, blockLen)

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

/*
func BenchmarkEvalFull(bench *testing.B) {
	// db parameters
	blockSize := 16
	numColumns := 200

	// generate random alpha
	alpha, err := new(field.Element).SetRandom(utils.RandomPRG())
	if err != nil {
		panic(err)
	}

	beta := make([]field.Element, blockSize+1)
	beta[0] = field.One()
	for i := 1; i < len(beta); i++ {
		beta[i].Mul(&beta[i-1], alpha)
	}

	// generate one key
	logN := uint64(bits.Len(uint(numColumns) - 1))
	key, _ := Gen(1, beta, logN)

	q := make([]field.Element, numColumns*(blockSize+1))

	bench.ResetTimer()
	bench.ReportAllocs()
	for i := 0; i < bench.N; i++ {
		EvalFullFlatten(key, logN, blockSize+1, q)
	}
}

func BenchmarkAES(b *testing.B) {
	prfkeyL := []byte{36, 156, 50, 234, 92, 230, 49, 9, 174, 170, 205, 160, 98, 236, 29, 243}
	keyL := make([]uint32, 11*4)
	expandKeyAsm(&prfkeyL[0], &keyL[0])
	dst := new(block)
	src := new(block)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		aes128MMO(&keyL[0], &dst[0], &src[0])
	}
}

/*
func BenchmarkXor16(bench *testing.B) {
	a := new(block)
	b := new(block)
	c := new(block)
	for i := 0; i < bench.N; i++ {
		xor16(&c[0], &b[0], &a[0])
	}
}

func TestEval(test *testing.T) {
	logN := uint64(8)
	alpha := uint64(123)
	beta := make([]field.Element, 2)
	beta[0].SetUint64(7613)
	beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

	sum := make([]field.Element, 2)
	out0 := make([]field.Element, 2)
	out1 := make([]field.Element, 2)
	zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		Eval(a, i, logN, out0)
		Eval(b, i, logN, out1)

		for j := 0; j < 2; j++ {
			sum[j].Add(&out0[j], &out1[j])
		}

		//log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
		if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
			test.Fail()
		}

		if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}

func TestEvalFull(test *testing.T) {
	logN := uint64(9)
	alpha := uint64(123)
	beta := make([]field.Element, 2)
	beta[0].SetUint64(7613)
	beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

	sum := make([]field.Element, 2)
	out0 := make([][]field.Element, 1<<logN)
	out1 := make([][]field.Element, 1<<logN)

	for i := 0; i < len(out0); i++ {
		out0[i] = make([]field.Element, 2)
		out1[i] = make([]field.Element, 2)
	}

	EvalFull(a, logN, out0)
	EvalFull(b, logN, out1)

	zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		for j := 0; j < 2; j++ {
			sum[j].Add(&out0[i][j], &out1[i][j])
		}

		//log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
		if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
			test.Fail()
		}

		if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}

func TestEvalFullShort(test *testing.T) {
	logN := uint64(2)
	alpha := uint64(2)
	beta := make([]field.Element, 2)
	beta[0].SetUint64(7613)
	beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

	sum := make([]field.Element, 2)
	out0 := make([][]field.Element, 1<<logN)
	out1 := make([][]field.Element, 1<<logN)

	for i := 0; i < len(out0); i++ {
		out0[i] = make([]field.Element, 2)
		out1[i] = make([]field.Element, 2)
	}

	EvalFull(a, logN, out0)
	EvalFull(b, logN, out1)

	zero := field.Zero()
	for i := uint64(0); i < (uint64(1) << logN); i++ {
		for j := 0; j < 2; j++ {
			sum[j].Add(&out0[i][j], &out1[i][j])
		}

		//log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
		if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
			test.Fail()
		}

		if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}

func TestEvalFullPartial(test *testing.T) {
	logN := uint64(9)
	alpha := uint64(123)
	beta := make([]field.Element, 2)
	beta[0].SetUint64(7613)
	beta[1].SetUint64(991)

	a, b := Gen(alpha, beta, logN)

	sum := make([]field.Element, 2)

	outlen := 278
	out0 := make([][]field.Element, outlen)
	out1 := make([][]field.Element, outlen)

	for i := 0; i < len(out0); i++ {
		out0[i] = make([]field.Element, 2)
		out1[i] = make([]field.Element, 2)
	}

	EvalFull(a, logN, out0)
	EvalFull(b, logN, out1)

	zero := field.Zero()
	for i := uint64(0); i < uint64(outlen); i++ {
		for j := 0; j < 2; j++ {
			sum[j].Add(&out0[i][j], &out1[i][j])
		}

		//log.Printf("%v %v %v %v", i, alpha, beta.String(), sum.String())
		if i != alpha && (!sum[0].Equal(&zero) || !sum[1].Equal(&zero)) {
			test.Fail()
		}

		if i == alpha && (!sum[0].Equal(&beta[0]) || !sum[1].Equal(&beta[1])) {
			test.Fail()
		}
	}
}
*/
