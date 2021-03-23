package server

import (
	"math"
	"runtime"
	"sync"

	"github.com/lukechampine/fastxor"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
)

// Server is a scheme-agnostic VPIR server interface, implemented by both IT
// and DPF-based schemes
type Server interface {
	AnswerBytes([]byte) ([]byte, error)
	DBInfo() *database.Info
}

/*
%%	VPIR primitives
*/

// Answer computes the VPIR answer for the given query
func answer(q []field.Element, db *database.DB, NGoRoutines int) []field.Element {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		a := make([]field.Element, db.NumRows)
		for i := 0; i < db.NumRows; i++ {
			for j := 0; j < db.NumColumns; j++ {
				if db.Entries[i*db.NumColumns+j].Equal(&cst.One) {
					a[i].Add(&a[i], &q[j])
				}
			}
		}
		return a
	}

	// %%% Logic %%%
	// compute the matrix-vector inner products,
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	// If numRows == 1, the db is a vector so we split it by giving columns to workers.
	// Otherwise, if the db is a matrix, we split by rows and give a chunk of rows to each worker.
	// The goal is to have a fixed number of workers and start them only once.
	var begin, end int
	if db.NumRows == 1 {
		columnsPerRoutine := utils.DivideAndRoundUpToMultiple(db.NumColumns, NGoRoutines, 1)
		// a channel to pass results from the routines back
		resultsChan := make(chan []field.Element, NGoRoutines*(db.BlockSize+1))
		numWorkers := 0
		// we need to traverse column by column
		for j := 0; j < db.NumColumns; j += columnsPerRoutine {
			columnsPerRoutine, begin, end = computeChunkIndices(j, columnsPerRoutine, db.BlockSize, db.NumColumns)
			go processColumns(db.Entries[begin:end], q[j*(db.BlockSize+1):(j+columnsPerRoutine)*(db.BlockSize+1)], db.BlockSize, resultsChan)
			numWorkers++
		}
		m := combineColumnResults(numWorkers, db.BlockSize+1, resultsChan)
		close(resultsChan)

		return m
	} else {
		m := make([]field.Element, db.NumRows*(db.BlockSize+1))
		var workers sync.WaitGroup
		rowsPerRoutine := utils.DivideAndRoundUpToMultiple(db.NumRows, NGoRoutines, 1)
		for j := 0; j < db.NumRows; j += rowsPerRoutine {
			rowsPerRoutine, begin, end = computeChunkIndices(j, rowsPerRoutine, db.BlockSize, db.NumRows)
			workers.Add(1)
			go processRows(m[j*(db.BlockSize+1):(j+rowsPerRoutine)*(db.BlockSize+1)],
				db.Entries[begin*db.NumColumns:end*db.NumColumns], q, &workers, db.NumColumns, db.BlockSize)
		}
		workers.Wait()

		return m
	}
}

// processing multiple rows by iterating over them
func processRows(output, rows, query []field.Element, wg *sync.WaitGroup, numColumns, blockLen int) {
	numElementsInRow := blockLen * numColumns
	for i := 0; i < len(rows)/numElementsInRow; i++ {
		res := computeMessageAndTag(rows[i*numElementsInRow:(i+1)*numElementsInRow], query, blockLen)
		copy(output[i*(blockLen+1):(i+1)*(blockLen+1)], res)
	}
	wg.Done()
}

// processing a chunk of a database row
func processColumns(columns, query []field.Element, blockLen int, reply chan<- []field.Element) {
	reply <- computeMessageAndTag(columns, query, blockLen)
}

// combine the results of processing a row by different routines
func combineColumnResults(nWrk int, resLen int, workerReplies <-chan []field.Element) []field.Element {
	product := make([]field.Element, resLen)
	for i := 0; i < nWrk; i++ {
		reply := <-workerReplies
		for i, elem := range reply {
			product[i].Add(&product[i], &elem)
		}
	}
	return product
}

// computeMessageAndTag multiplies db entries with the elements
// from the client query and computes a tag over each block
func computeMessageAndTag(elements, q []field.Element, blockLen int) []field.Element {
	var prodTag, prod field.Element
	sumTag := field.Zero()
	sum := field.ZeroVector(blockLen)
	for j := 0; j < len(elements)/blockLen; j++ {
		for b := 0; b < blockLen; b++ {
			if elements[j*blockLen+b].IsZero() {
				// no need to multiply if the element value is zero
				continue
			}
			// compute message
			prod.Mul(&elements[j*blockLen+b], &q[j*(blockLen+1)])
			sum[b].Add(&sum[b], &prod)
			// compute block tag
			prodTag.Mul(&elements[j*blockLen+b], &q[j*(blockLen+1)+1+b])
			sumTag.Add(&sumTag, &prodTag)
		}
	}
	return append(sum, sumTag)
}

/*
%%	PIR primitives
*/
func answerPIR(q []byte, db *database.Bytes) []byte {
	m := make([]byte, db.NumRows*db.BlockSize)
	// multithreading
	NGoRoutines := runtime.NumCPU()
	var begin, end int
	// Vector db
	if db.NumRows == 1 {
		columnsPerRoutine := utils.DivideAndRoundUpToMultiple(db.NumColumns, NGoRoutines, 8)
		// a channel to pass results from the routines back
		resultsChan := make(chan []byte, NGoRoutines*db.BlockSize)
		numWorkers := 0
		for j := 0; j < db.NumColumns; j += columnsPerRoutine {
			columnsPerRoutine, begin, end = computeChunkIndices(j, columnsPerRoutine, db.BlockSize, db.NumColumns)
			// We need /8 because q is packed with 1 bit per block
			go xorColumns(db.Entries[begin:end], q[j/8:int(math.Ceil(float64(j+columnsPerRoutine)/8))], db.BlockSize, resultsChan)
			numWorkers++
		}
		m = combineColumnXORs(numWorkers, db.BlockSize, resultsChan)
		close(resultsChan)
		return m
	} else {
		//	Matrix db
		var workers sync.WaitGroup
		rowsPerRoutine := utils.DivideAndRoundUpToMultiple(db.NumRows, NGoRoutines, 1)
		for j := 0; j < db.NumRows; j += rowsPerRoutine {
			rowsPerRoutine, begin, end = computeChunkIndices(j, rowsPerRoutine, db.BlockSize, db.NumRows)
			workers.Add(1)
			go xorRows(m[begin:end], db.Entries[begin*db.NumColumns:end*db.NumColumns], q, &workers, db.NumColumns, db.BlockSize)
		}
		workers.Wait()

		return m
	}
}

// XORs entries and q block by block of size bl
func xorValues(entries, q []byte, bl int) []byte {
	sum := make([]byte, bl)
	for j := 0; j < len(entries)/bl; j++ {
		if (q[j/8]>>(j%8))&1 == byte(1) {
			fastxor.Bytes(sum, sum, entries[j*bl:(j+1)*bl])
		}
	}
	return sum
}

// XORs columns in the same row
func xorColumns(columns, query []byte, blockLen int, reply chan<- []byte) {
	reply <- xorValues(columns, query, blockLen)
}

// XORs all the columns in a row, row by row, and writes the result into output
func xorRows(output, rows, query []byte, wg *sync.WaitGroup, numColumns, blockLen int) {
	numElementsInRow := blockLen * numColumns
	for i := 0; i < len(rows)/numElementsInRow; i++ {
		res := xorValues(rows[i*numElementsInRow:(i+1)*numElementsInRow], query, blockLen)
		copy(output[i*blockLen:(i+1)*blockLen], res)
	}
	wg.Done()
}

// Waits for column XORs from individual workers and XORs the results together
func combineColumnXORs(nWrk int, blockLen int, workerReplies <-chan []byte) []byte {
	sum := make([]byte, blockLen)
	for i := 0; i < nWrk; i++ {
		reply := <-workerReplies
		fastxor.Bytes(sum, sum, reply)
	}
	return sum
}

/*
%%	Shared helpers
*/
func computeChunkIndices(ind, step, multiplier, max int) (int, int, int) {
	// avoiding overflow when colPerChunk does not divide db.Columns evenly
	if ind+step > max {
		step = max - ind
	}
	return step, ind * multiplier, (ind + step) * multiplier
}
