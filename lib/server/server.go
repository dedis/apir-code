package server

import (
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
	// a channel to pass results from the routines back
	replies := make([]chan []field.Element, NGoRoutines)
	// Vector db
	if db.NumRows == 1 {
		columnsPerRoutine := db.NumColumns / NGoRoutines
		for i := 0; i < NGoRoutines; i++ {
			begin, end = computeChunkIndices(i, columnsPerRoutine, NGoRoutines-1, db.NumColumns)
			replyChan := make(chan []field.Element, db.BlockSize+1)
			replies[i] = replyChan
			go processColumns(db.Range(begin*db.BlockSize, end*db.BlockSize), q[begin*(db.BlockSize+1):end*(db.BlockSize+1)], db.BlockSize, replyChan)
		}
		m := make([]field.Element, db.BlockSize+1)
		for i, reply := range replies {
			chunk := <-reply
			for i, elem := range chunk {
				m[i].Add(&m[i], &elem)
			}
			close(replies[i])
		}
		return m
	} else {
		//	Matrix db
		rowsPerRoutine := db.NumRows / NGoRoutines
		for i := 0; i < NGoRoutines; i++ {
			begin, end = computeChunkIndices(i, rowsPerRoutine, NGoRoutines-1, db.NumRows)
			replyChan := make(chan []field.Element, (end-begin)*(db.BlockSize+1))
			replies[i] = replyChan
			go processRows(db.Range(begin*db.NumColumns*db.BlockSize, end*db.NumColumns*db.BlockSize), q, db.NumColumns, db.BlockSize, replyChan)
		}
		m := make([]field.Element, 0, db.NumRows*(db.BlockSize+1))
		for i, reply := range replies {
			chunk := <-reply
			m = append(m, chunk...)
			close(replies[i])
		}
		return m
	}
}

func answerNew(q field.ElemSlice, db *database.DB, NGoRoutines int) []field.Element {
	// Doing simplified scheme if block consists of a single bit
	if db.BlockSize == cst.SingleBitBlockLength {
		panic("not implemented")
	}

	// %%% Logic %%%
	// compute the matrix-vector inner products,
	// addition and multiplication of elements
	// in DB(2^128)^b are executed component-wise
	// If numRows == 1, the db is a vector so we split it by giving columns to workers.
	// Otherwise, if the db is a matrix, we split by rows and give a chunk of rows to each worker.
	// The goal is to have a fixed number of workers and start them only once.
	var begin, end int
	// a channel to pass results from the routines back
	replies := make([]chan []field.Element, NGoRoutines)

	// Vector db
	if db.NumRows == 1 {
		columnsPerRoutine := db.NumColumns / NGoRoutines
		for i := 0; i < NGoRoutines; i++ {
			begin, end = computeChunkIndices(i, columnsPerRoutine, NGoRoutines-1, db.NumColumns)
			replyChan := make(chan []field.Element, db.BlockSize+1)
			replies[i] = replyChan
			go processColumnsNew(db.Range(begin*db.BlockSize, end*db.BlockSize), q.Range(begin*(db.BlockSize+1), end*(db.BlockSize+1)), db.BlockSize, replyChan)
		}
		m := make([]field.Element, db.BlockSize+1)
		for i, reply := range replies {
			chunk := <-reply
			for i, elem := range chunk {
				m[i].Add(&m[i], &elem)
			}
			close(replies[i])
		}
		return m
	}

	//	Matrix db
	rowsPerRoutine := db.NumRows / NGoRoutines
	for i := 0; i < NGoRoutines; i++ {
		begin, end = computeChunkIndices(i, rowsPerRoutine, NGoRoutines-1, db.NumRows)
		replyChan := make(chan []field.Element, (end-begin)*(db.BlockSize+1))
		replies[i] = replyChan
		go processRowsNew(db.Range(begin*db.NumColumns*db.BlockSize, end*db.NumColumns*db.BlockSize), q, db.NumColumns, db.BlockSize, replyChan)
	}
	m := make([]field.Element, 0, db.NumRows*(db.BlockSize+1))
	for i, reply := range replies {
		chunk := <-reply
		m = append(m, chunk...)
		close(replies[i])
	}
	return m

}

// processing multiple rows by iterating over them
func processRows(rows database.ElementRange, query []field.Element, numColumns, blockLen int, reply chan<- []field.Element) {
	numElementsInRow := blockLen * numColumns
	numRowsToProcess := rows.Len() / numElementsInRow
	sums := make([]field.Element, 0, numRowsToProcess*(blockLen+1))
	for i := 0; i < numRowsToProcess; i++ {
		res := computeMessageAndTag(rows.Range(i*numElementsInRow, (i+1)*numElementsInRow), query, blockLen)
		sums = append(sums, res...)
	}
	reply <- sums
}

func processRowsNew(rows database.ElementRange, query field.ElemSlice, numColumns, blockLen int, reply chan<- []field.Element) {
	numElementsInRow := blockLen * numColumns
	numRowsToProcess := rows.Len() / numElementsInRow
	sums := make([]field.Element, 0, numRowsToProcess*(blockLen+1))
	for i := 0; i < numRowsToProcess; i++ {
		res := computeMessageAndTagNew(rows.Range(i*numElementsInRow, (i+1)*numElementsInRow), query, blockLen)
		sums = append(sums, res...)
	}
	reply <- sums
}

// processing a chunk of a database row
func processColumns(columns database.ElementRange, query []field.Element, blockLen int, reply chan<- []field.Element) {
	reply <- computeMessageAndTag(columns, query, blockLen)
}

func processColumnsNew(columns database.ElementRange, query field.ElemSlice, blockLen int, reply chan<- []field.Element) {
	reply <- computeMessageAndTagNew(columns, query, blockLen)
}

// computeMessageAndTag multiplies db entries with the elements
// from the client query and computes a tag over each block
func computeMessageAndTag(elements database.ElementRange, q []field.Element, blockLen int) []field.Element {
	var prodTag, prod field.Element
	sumTag := field.Zero()
	sum := field.ZeroVector(blockLen)
	for j := 0; j < elements.Len()/blockLen; j++ {
		for b := 0; b < blockLen; b++ {
			e := elements.Get(j*blockLen + b)
			if e.IsZero() {
				// no need to multiply if the element value is zero
				continue
			}
			// compute message
			e = elements.Get(j*blockLen + b)
			prod.Mul(&e, &q[j*(blockLen+1)])
			sum[b].Add(&sum[b], &prod)
			// compute block tag
			e = elements.Get(j*blockLen + b)
			prodTag.Mul(&e, &q[j*(blockLen+1)+1+b])
			sumTag.Add(&sumTag, &prodTag)
		}
	}
	return append(sum, sumTag)
}

func computeMessageAndTagNew(elements database.ElementRange, q field.ElemSlice, blockLen int) []field.Element {
	var prodTag, prod field.Element
	sumTag := field.Zero()
	sum := field.ZeroVector(blockLen)
	for j := 0; j < elements.Len()/blockLen; j++ {
		for b := 0; b < blockLen; b++ {
			e := elements.Get(j*blockLen + b)
			if e.IsZero() {
				// no need to multiply if the element value is zero
				continue
			}
			// compute message
			e = elements.Get(j*blockLen + b)
			g := q.Get(j * (blockLen + 1))
			prod.Mul(&e, &g)
			sum[b].Add(&sum[b], &prod)
			// compute block tag
			e = elements.Get(j*blockLen + b)
			f := q.Get(j*(blockLen+1) + 1 + b)
			prodTag.Mul(&e, &f)
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

func processSingleBitColumnsNew(begin, end int, db *database.DB, q field.ElemSlice, replyTo chan<- field.Element) {
	reply := field.Zero()
	for j := begin; j < end; j++ {
		entry := db.GetEntry(j)
		if entry.Equal(&cst.One) {
			e := q.Get(j)
			reply.Add(&reply, &e)
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
		// Divide and multiple by 8 to make sure that
		// each routine process whole bytes from the query, i.e.,
		// a byte does not get split between different routines
		columnsPerRoutine := ((db.NumColumns / NGoRoutines) / 8) * 8
		for i := 0; i < NGoRoutines; i++ {
			begin, end = computeChunkIndices(i, columnsPerRoutine, NGoRoutines-1, db.NumColumns)
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
			begin, end = computeChunkIndices(i, rowsPerRoutine, NGoRoutines-1, db.NumRows)
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
func computeChunkIndices(ind, multiplier, maxIndex, maxValue int) (int, int) {
	begin, end := ind*multiplier, (ind+1)*multiplier
	// the last routine takes all the left-overs
	if ind == maxIndex {
		end = maxValue
	}
	return begin, end
}
