package server

import (
	"math"
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
		// make sure that we do not need up with routines processing 0 elements
		if NGoRoutines > db.NumColumns {
			NGoRoutines = db.NumColumns
		}
		columnsPerRoutine := db.NumColumns / NGoRoutines
		replies := make([]chan field.Element, NGoRoutines)
		m := make([]field.Element, db.NumRows)
		var begin, end int
		for i := 0; i < db.NumRows; i++ {
			for j := 0; j < NGoRoutines; j++ {
				begin, end = j*columnsPerRoutine, (j+1)*columnsPerRoutine
				// make the last routine take all the left-over (from division) columns
				if j == NGoRoutines-1 {
					end = db.NumColumns
				}
				replyChan := make(chan field.Element, 1)
				replies[i] = replyChan
				go processSingleBitColumns(begin, end, db, q, replyChan)
			}

			m[i].SetZero()
			for j, reply := range replies {
				element := <-reply
				m[i].Add(&m[i], &element)
				close(replies[j])
			}
		}
		return m
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

			go processColumns(db.Range(begin, end), q[j*(db.BlockSize+1):(j+columnsPerRoutine)*(db.BlockSize+1)], db.BlockSize, resultsChan)
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
				db.Range(begin*db.NumColumns, end*db.NumColumns), q, &workers, db.NumColumns, db.BlockSize)
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

// Processing of columns for a database where each field element
// encodes just a single bit
func processSingleBitColumns(begin, end int, db *database.DB, q []field.Element, replyTo chan<- field.Element) {
	reply := field.Zero()
	for j := begin; j < end; j++ {
		entry := db.GetEntry(j)
		if entry.Equal(&cst.One) {
			reply.Add(&reply, &q[j])
		}
	}
	replyTo <- reply
}

/*
%%	PIR primitives
*/
func answerPIR(q []byte, db *database.Bytes, NGoRoutines int) []byte {
	var begin, end int
	// a channel to pass results from the routines back
	replies := make([]chan []byte, NGoRoutines)
	// Vector db
	if db.NumRows == 1 {
		columnsPerRoutine := ((db.NumColumns / NGoRoutines) / 8) * 8
		for i := 0; i < NGoRoutines; i++ {
			begin, end = i*columnsPerRoutine, (i+1)*columnsPerRoutine
			// the last routine takes all the left-overs
			if i == NGoRoutines-1 {
				end = db.NumColumns
			}
			replyChan := make(chan []byte, db.BlockSize)
			replies[i] = replyChan
			// We need /8 because q is packed with 1 bit per block
			go xorColumns(db.Entries[begin*db.BlockSize:end*db.BlockSize], q[begin/8:int(math.Ceil(float64(end)/8))], db.BlockSize, replyChan)
		}
		m := make([]byte, db.BlockSize)
		for i, reply := range replies {
			chunk := <-reply
			fastxor.Bytes(m, m, chunk)
			close(replies[i])
		}
		return m
	} else {
		//	Matrix db
		rowsPerRoutine := db.NumRows / NGoRoutines
		for i := 0; i < NGoRoutines; i++ {
			begin, end = i*rowsPerRoutine, (i+1)*rowsPerRoutine
			// the last routine takes all the left-overs
			if i == NGoRoutines-1 {
				end = db.NumRows
			}
			replyChan := make(chan []byte, (end-begin)*db.BlockSize)
			replies[i] = replyChan
			go xorRows(db.Entries[begin*db.NumColumns*db.BlockSize:end*db.NumColumns*db.BlockSize], q, db.NumColumns, db.BlockSize, replyChan)
		}
		m := make([]byte, 0, db.NumRows*db.BlockSize)
		for i, reply := range replies {
			chunk := <-reply
			m = append(m, chunk...)
			close(replies[i])
		}
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
func xorRows(rows, query []byte, numColumns, blockLen int, reply chan<- []byte) {
	numElementsInRow := blockLen * numColumns
	numRowsToProcess := len(rows) / numElementsInRow
	sums := make([]byte, 0, numRowsToProcess*blockLen)
	for i := 0; i < numRowsToProcess; i++ {
		res := xorValues(rows[i*numElementsInRow:(i+1)*numElementsInRow], query, blockLen)
		sums = append(sums, res...)
	}
	reply <- sums
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
