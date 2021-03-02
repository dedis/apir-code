package server

import (
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"runtime"
	"sync"
)

// Server is a scheme-agnostic VPIR server interface, implemented by both IT
// and DPF-based schemes
type Server interface {
	AnswerBytes([]byte) ([]byte, error)
	DBInfo() *database.Info
}

// Answer computes the answer for the given query
func answer(q []field.Element, db *database.DB) []field.Element {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		a := make([]field.Element, db.NumRows)
		for i := 0; i < db.NumRows; i++ {
			for j := 0; j < db.NumColumns; j++ {
				if db.Entries[i][j].Equal(&cst.One) {
					a[i].Add(&a[i], &q[j])
				}
			}
		}
		return a
	}

	// parse the query
	qZeroBase := make([]field.Element, db.NumColumns)
	qOne := make([]field.Element, db.NumColumns*db.BlockSize)
	for j := 0; j < db.NumColumns; j++ {
		qZeroBase[j] = q[j*(db.BlockSize+1)]
		copy(qOne[j*db.BlockSize:(j+1)*db.BlockSize], q[j*(db.BlockSize+1)+1:(j+1)*(db.BlockSize+1)])
	}

	// multithreading
	var wg sync.WaitGroup
	// channel to pass the results from the routines back
	results := make([]chan []field.Element, db.NumRows)
	numCPU := runtime.NumCPU()
	var colChunkLen int = db.NumColumns / numCPU
	// compute the matrix-vector inner products
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	m := make([]field.Element, db.NumRows*(db.BlockSize+1))
	// we have to traverse column by column
	for i := 0; i < db.NumRows; i++ {
		for c := 0; c < numCPU; c++ {
			wg.Add(1)
			go processChunk(db.Entries[i][c*colChunkLen*db.BlockSize:(c+1)*colChunkLen*db.BlockSize], db.BlockSize,
				qZeroBase[c*colChunkLen:(c+1)*colChunkLen], qOne[c*colChunkLen*db.BlockSize:(c+1)*colChunkLen*db.BlockSize], results[i], &wg)
		}
		go combineChunkResults(m[i*(db.BlockSize+1):(i+1)*(db.BlockSize+1)], results[i])
	}
	wg.Wait()
	for i := range results {
		close(results[i])
	}

	return m
}

func processChunk(dbChunk []field.Element, blockLen int, qZ []field.Element, qO []field.Element,
																	reply chan<- []field.Element, wg *sync.WaitGroup) {
	var prodTag, prod field.Element
	sumTag := field.Zero()
	sum := field.ZeroVector(blockLen)
	for j := 0; j < len(dbChunk)/blockLen; j++ {
		for b := 0; b < blockLen; b++ {
			if dbChunk[j*blockLen+b].IsZero() {
				// no need to multiply is the the element value is zero
				continue
			}
			prod.Mul(&dbChunk[j*blockLen+b], &qZ[j])
			sum[b].Add(&sum[b], &prod)

			// compute block tag
			prodTag.Mul(&dbChunk[j*blockLen+b], &qO[j*blockLen+b])
			sumTag.Add(&sumTag, &prodTag)
		}
	}
	reply <- append(sum, sumTag)
	wg.Done()
}

func combineChunkResults(product []field.Element, workerReplies <-chan []field.Element) {
	for reply := range workerReplies {
		for i, elem := range reply {
			product[i].Add(&product[i], &elem)
		}
	}
}
