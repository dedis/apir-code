package server

import (
	"fmt"
	"math"

	"github.com/lukechampine/fastxor"
	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
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
func answer(q []uint32, db *database.DB, NGoRoutines int) []uint32 {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		// make sure that we do not need up with routines processing 0 elements
		if NGoRoutines > db.NumColumns {
			NGoRoutines = db.NumColumns
		}
		columnsPerRoutine := db.NumColumns / NGoRoutines
		replies := make([]chan uint32, NGoRoutines)
		m := make([]uint32, db.NumRows)
		var begin, end int
		for i := 0; i < db.NumRows; i++ {
			for j := 0; j < NGoRoutines; j++ {
				begin, end = j*columnsPerRoutine, (j+1)*columnsPerRoutine
				// make the last routine take all the left-over (from division) columns
				if j == NGoRoutines-1 {
					end = db.NumColumns
				}
				replyChan := make(chan uint32)
				replies[j] = replyChan
				go processSingleBitColumns(db.Range(begin, end), q[begin:end], replyChan)
			}

			for j, reply := range replies {
				element := <-reply
				m[i] = (m[i] + element) % field.ModP
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
	var prevElemPos, nextElemPos int
	// a channel to pass results from the routines back
	replies := make([]chan []uint32, NGoRoutines)
	// Vector db
	if db.NumRows == 1 {
		columnsPerRoutine := db.NumColumns / NGoRoutines
		for i := 0; i < NGoRoutines; i++ {
			begin, end := computeChunkIndices(i, columnsPerRoutine, NGoRoutines-1, db.NumColumns)
			for colN := begin; colN < end; colN++ {
				nextElemPos += db.BlockLengths[colN]
			}
			replyChan := make(chan []uint32, db.BlockSize+1)
			replies[i] = replyChan
			go processColumns(db.Range(prevElemPos, nextElemPos), db.BlockLengths[begin:end], q[begin*(db.BlockSize+1):end*(db.BlockSize+1)], db.BlockSize, replyChan)
			prevElemPos = nextElemPos
		}
		m := make([]uint32, db.BlockSize+1)
		for i, reply := range replies {
			chunk := <-reply
			for i, elem := range chunk {
				m[i] = (m[i] + elem) % field.ModP
			}
			close(replies[i])
		}
		return m
	} else {
		//	Matrix db
		rowsPerRoutine := db.NumRows / NGoRoutines
		for i := 0; i < NGoRoutines; i++ {
			begin, end := computeChunkIndices(i, rowsPerRoutine, NGoRoutines-1, db.NumRows)
			for rowN := begin; rowN < end; rowN++ {
				for colN := 0; colN < db.NumColumns; colN++ {
					nextElemPos += db.BlockLengths[rowN*db.NumColumns+colN]
				}
			}
			replyChan := make(chan []uint32, (end-begin)*(db.BlockSize+1))
			replies[i] = replyChan
			go processRows(db.Range(prevElemPos, nextElemPos), db.BlockLengths[begin*db.NumColumns:end*db.NumColumns], q, end-begin, db.NumColumns, db.BlockSize, replyChan)
			prevElemPos = nextElemPos
		}
		m := make([]uint32, 0, db.NumRows*(db.BlockSize+1))
		for i, reply := range replies {
			chunk := <-reply
			m = append(m, chunk...)
			close(replies[i])
		}
		return m
	}
}

// processing multiple rows by iterating over them
func processRows(rows []uint32, blockLens []int, query []uint32, numRowsToProcess, numColumns, blockLen int, reply chan<- []uint32) {
	var prevPos, nextPos int
	sums := make([]uint32, 0, numRowsToProcess*(blockLen+1))
	for i := 0; i < numRowsToProcess; i++ {
		for j := 0; j < numColumns; j++ {
			nextPos += blockLens[i*numColumns+j]
		}
		res := computeMessageAndTag(rows[prevPos:nextPos], blockLens[i*numColumns:(i+1)*numColumns], query, blockLen)
		sums = append(sums, res...)
		prevPos = nextPos
	}
	reply <- sums
}

// processing a chunk of a database row
func processColumns(columns []uint32, blockLens []int, query []uint32, blockLen int, reply chan<- []uint32) {
	reply <- computeMessageAndTag(columns, blockLens, query, blockLen)
}

// computeMessageAndTag multiplies db entries with the elements
// from the client query and computes a tag over each block
func computeMessageAndTag(elements []uint32, blockLens []int, q []uint32, blockLen int) []uint32 {
	sumTag := uint32(0)
	sum := make([]uint32, blockLen) // already initilized at zero
	pos := 0
	for j := 0; j < len(blockLens); j++ {
		for b := 0; b < blockLens[j]; b++ {
			fmt.Println(elements)
			fmt.Println("len(elements):", len(elements))
			fmt.Println(pos)
			fmt.Println(elements[pos])
			if elements[pos] == 0 {
				// no need to multiply if the element value is zero
				pos += 1
				continue
			}
			// compute message
			prod := (uint64(elements[pos]) * uint64(q[j*(blockLen+1)])) % uint64(field.ModP)
			sum[b] = (sum[b] + uint32(prod)) % field.ModP
			// compute block tag
			prodTag := (uint64(elements[pos]) * uint64(q[j*(blockLen+1)+1+b])) % uint64(field.ModP)
			sumTag = (sumTag + uint32(prodTag)) % field.ModP
			pos += 1
		}
	}
	return append(sum, sumTag)
}

// Processing of columns for a database where each field element
// encodes just a single bit
func processSingleBitColumns(elements []uint32, q []uint32, replyTo chan<- uint32) {
	reply := uint32(0)
	for j := 0; j < len(elements); j++ {
		if elements[j] == 1 {
			reply = (reply + q[j]) % field.ModP
		}
	}
	replyTo <- reply
}

/*
%%	PIR primitives
*/
func answerPIR(q []byte, db *database.Bytes, NGoRoutines int) []byte {
	// a channel to pass results from the routines back
	replies := make([]chan []byte, NGoRoutines)
	// Vector db
	var prevElemPos, nextElemPos int
	if db.NumRows == 1 {
		// Divide and multiple by 8 to make sure that
		// each routine process whole bytes from the query, i.e.,
		// a byte does not get split between different routines
		columnsPerRoutine := ((db.NumColumns / NGoRoutines) / 8) * 8
		for i := 0; i < NGoRoutines; i++ {
			begin, end := computeChunkIndices(i, columnsPerRoutine, NGoRoutines-1, db.NumColumns)
			for colN := begin; colN < end; colN++ {
				nextElemPos += db.BlockLengths[colN]
			}
			replyChan := make(chan []byte, db.BlockSize)
			replies[i] = replyChan
			// We need /8 because q is packed with 1 bit per block
			go xorColumns(db.Entries[prevElemPos:nextElemPos], db.BlockLengths[begin:end], q[begin/8:int(math.Ceil(float64(end)/8))], db.BlockSize, replyChan)
			prevElemPos = nextElemPos
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
			begin, end := computeChunkIndices(i, rowsPerRoutine, NGoRoutines-1, db.NumRows)
			for rowN := begin; rowN < end; rowN++ {
				for colN := 0; colN < db.NumColumns; colN++ {
					nextElemPos += db.BlockLengths[rowN*db.NumColumns+colN]
				}
			}
			replyChan := make(chan []byte, (end-begin)*db.BlockSize)
			replies[i] = replyChan
			go xorRows(db.Entries[prevElemPos:nextElemPos], db.BlockLengths[begin*db.NumColumns:end*db.NumColumns], q, end-begin, db.NumColumns, db.BlockSize, replyChan)
			prevElemPos = nextElemPos
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
func xorValues(entries []byte, blockLens []int, q []byte, bl int) []byte {
	sum := make([]byte, bl)
	pos := 0
	for j := range blockLens {
		if (q[j/8]>>(j%8))&1 == byte(1) {
			fastxor.Bytes(sum, sum, entries[pos:pos+blockLens[j]])
		}
		pos += blockLens[j]
	}
	return sum
}

// XORs columns in the same row
func xorColumns(columns []byte, blockLens []int, query []byte, blockLen int, reply chan<- []byte) {
	reply <- xorValues(columns, blockLens, query, blockLen)
}

// XORs all the columns in a row, row by row, and writes the result into output
func xorRows(rows []byte, blockLens []int, query []byte, numRowsToProcess, numColumns, blockLen int, reply chan<- []byte) {
	var prevPos, nextPos int
	sums := make([]byte, 0, numRowsToProcess*blockLen)
	for i := 0; i < numRowsToProcess; i++ {
		for j := 0; j < numColumns; j++ {
			nextPos += blockLens[i*numColumns+j]
		}
		res := xorValues(rows[prevPos:nextPos], blockLens[i*numColumns:(i+1)*numColumns], query, blockLen)
		sums = append(sums, res...)
		prevPos = nextPos
	}
	reply <- sums
}

/*
%%	Shared helpers
*/
func computeChunkIndices(ind, multiplier, maxIndex, maxValue int) (int, int) {
	begin, end := ind*multiplier, (ind+1)*multiplier
	// the last routine takes all the left-overs
	if ind == maxIndex {
		end = maxValue
	}
	return begin, end
}
