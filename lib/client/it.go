package client

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"golang.org/x/xerrors"
)

// Information theoretic client for single-bit and multi-bit schemes
// working in F(2^127-1). Both vector and matrix (rebalanced)
// representations of the database are handled by this client, via
// a boolean variable

// IT represents the client for the information theoretic multi-bit scheme
type IT struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewIT returns a client for the information theoretic multi-bit
// scheme, working both with the vector and the rebalanced representation of
// the database.
func NewIT(rnd io.Reader, info *database.Info) *IT {
	return &IT{
		// 👉 got a panic with the old rand reader
		rnd:    rand.Reader,
		dbInfo: info,
		state:  nil,
	}
}

func (c *IT) QueryBytes(index, numServers int) ([][]byte, error) {
	panic("not implemented")
}

func (c *IT) QueryBytesNew(index, numServers int) (*BatchIterator, error) {
	// get reconstruction
	queries := c.Query(index, numServers)

	// 👉 the old way: getting all results and then encoding it. At first it was
	// using gob encoder, which uses a lot of memory. Then I used binary

	// encode all the queries in bytes
	// out := make([][]byte, len(queries))
	// for i := range queries {
	// 	// queryBuf := make([]byte, len(queries[i])*8*2)
	// 	// for k := 0; k < len(queries[i]); k++ {
	// 	// 	binary.LittleEndian.PutUint64(queryBuf[k*8*2:k*8*2+8], queries[i][k][0])
	// 	// 	binary.LittleEndian.PutUint64(queryBuf[k*8*2+8:k*8*2+8+8], queries[i][k][1])
	// 	// }
	// 	// out[i] = queryBuf
	// 	out[i] = queries[i].Bytes()

	// 	// buf := new(bytes.Buffer)
	// 	// enc := gob.NewEncoder(buf)
	// 	// if err := enc.Encode(queries[i]); err != nil {
	// 	// 	return nil, err
	// 	// }
	// 	// out[i] = buf.Bytes()
	// }

	return queries, nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *IT) Query(index, numServers int) *BatchIterator {
	if invalidQueryInputsIT(index, numServers) {
		log.Fatal("invalid query inputs")
	}

	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	if err != nil {
		log.Fatal(err)
	}

	// vectors, err := c.secretShare(numServers)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// 👉 the only things we need to keep in memory during the batch processsing
	// is the result of the last columns, to which all other columns accumulate
	// their result, and the current column.
	lastCol := field.NewElemSlice(len(c.state.a))
	lastCol.SetRandom(rand.Reader)

	result := NewBatchIterator(c.dbInfo.NumColumns, len(c.state.a), c.state.iy, numServers, c.state)

	// 👉 another way to iterate over the results, which required to all results
	// of all columns at once, and no batch.

	// iter := newQueryIterator(c.dbInfo.NumColumns, numServers, len(c.state.a),
	// 	c.state.iy, c.rnd, c.state)

	return result
}

// ReconstructBytes returns []field.Element
func (c *IT) ReconstructBytes(a [][]byte) (interface{}, error) {
	// answer, err := decodeAnswer(a)
	// if err != nil {
	// 	return nil, err
	// }

	// 👉 we're using binary encoding, which is more memory efficient.
	res := make([][]field.Element, len(a))

	for i := range res {
		n := len(a[i]) / (8 * 2)
		data := make([]field.Element, n)

		for k := 0; k < n; k++ {
			memIndex := k * 8 * 2

			data[k] = field.Element{
				binary.LittleEndian.Uint64(a[i][memIndex : memIndex+8]),
				binary.LittleEndian.Uint64(a[i][memIndex+8 : memIndex+16]),
			}
		}

		res[i] = data
	}

	return c.Reconstruct(res)
}

func (c *IT) Reconstruct(answers [][]field.Element) ([]field.Element, error) {
	return reconstruct(answers, c.dbInfo, c.state)
}

// secretShare the vector a among numServers non-colluding servers
func (c *IT) secretShare(numServers int) (*field.ElemSliceIterator, error) {
	// get block length
	blockLen := len(c.state.a)
	// Number of field elements in the whole vector
	vectorLen := c.dbInfo.NumColumns * blockLen

	result := make([]field.ElemSlice, numServers)
	for k := range result {
		result[k] = field.NewElemSlice(vectorLen)

		err := result[k].SetRandom(c.rnd)
		if err != nil {
			return nil, xerrors.Errorf("failed to set random: %v", err)
		}
	}

	// create query vectors for all the servers F^(1+b)
	// vectors := make([][]field.Element, numServers)
	// for k := range vectors {
	// 	vectors[k] = make([]field.Element, vectorLen)
	// }

	// Get random elements for all numServers-1 vectors
	// rand, err := field.RandomVector(c.rnd, (numServers-1)*vectorLen)
	// if err != nil {
	// 	return nil, err
	// }
	// perform additive secret sharing
	var colStart, colEnd int
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		colStart = j * blockLen
		colEnd = (j + 1) * blockLen
		// Assign k - 1 random vectors of length dbLength containing
		// elements in F^(1+b)
		// for k := 0; k < numServers-1; k++ {
		// rand, err := field.RandomVector(c.rnd, colEnd-colStart)
		// if err != nil {
		// 	panic(err)
		// }
		// copy(vectors[k][colStart:colEnd], rand)
		// copy(vectors[k][colStart:colEnd], rand[k*vectorLen+colStart:k*vectorLen+colEnd])
		// }

		// we should perform component-wise additive secret sharing
		for b := colStart; b < colEnd; b++ {
			sum := field.Zero()
			for k := 0; k < numServers-1; k++ {
				a := result[k].Get(b)
				sum.Add(&sum, &a)
			}
			// vectors[numServers-1][b].Set(&sum)
			// vectors[numServers-1][b].Neg(&vectors[numServers-1][b])

			sum.Neg(&sum)

			// set alpha vector at the block we want to retrieve
			if j == c.state.iy {
				// vectors[numServers-1][b].Add(&vectors[numServers-1][b], &c.state.a[b-j*blockLen])
				sum.Add(&sum, &c.state.a[b-j*blockLen])

			}

			result[numServers-1].Set(b, sum)
		}
	}

	// return vectors, nil
	return field.NewElemSliceIterator(result), nil
}

// NewBatchIterator creates a new batch iterator
func NewBatchIterator(totalCols, colSize, iy, numServers int, s *state) *BatchIterator {
	return &BatchIterator{
		currentCol: 0,
		totalCols:  totalCols,
		colSize:    colSize,
		iy:         iy,
		numServers: numServers,

		state: s,
	}
}

// BatchIterator iterates over the columns, and offers an iterator of servers
// for each batch.
//
// [BATCH A] -> Server 1
//           -> Server 2
// [BATCH B] -> Server 1
//           -> Server 2
//  ...
type BatchIterator struct {
	currentCol int
	totalCols  int
	colSize    int
	iy         int
	numServers int

	state *state
}

// HasNext tells it there is additional batches.
func (b *BatchIterator) HasNext() bool {
	return b.currentCol < b.totalCols
}

// GetNext return the next batch. Be sure that HasNext returns true.
func (b *BatchIterator) GetNext(numCols int) *QueriesIterator {
	if b.currentCol+numCols >= b.totalCols {
		numCols = b.totalCols - b.currentCol
	}

	// with the last column, each server except the last one will add its result
	// to it: lastCol[i] = colSrv1[i] + colSrv2[i] + ...
	lastCol := field.NewElemSlice(b.colSize * numCols)

	result := NewQueriesIterator(b.numServers, numCols, b.colSize, b.currentCol,
		b.iy, b.state, &lastCol)

	b.currentCol += numCols

	return result
}

// NewQueriesIterator returns a new server iterator
func NewQueriesIterator(numServers, numCols, colSize, currentCol, iy int, s *state,
	lastCol *field.ElemSlice) *QueriesIterator {

	return &QueriesIterator{
		numServers:    numServers,
		currentServer: 0,
		currentCol:    currentCol,
		numCols:       numCols,
		colSize:       colSize,
		iy:            iy,

		state:   s,
		lastCol: lastCol,
	}
}

// QueriesIterator is an iterator for queries in a given batch for each server:
// -> batch server 1
// -> batch server 2
// -> ...
type QueriesIterator struct {
	currentServer int
	numServers    int
	currentCol    int
	numCols       int
	colSize       int
	iy            int

	state *state

	lastCol *field.ElemSlice
}

// HasNext returns if there are additional query
func (s *QueriesIterator) HasNext() bool {
	return s.currentServer < s.numServers
}

// GetNext return the next query for the next server
func (s *QueriesIterator) GetNext() field.ElemSlice {

	// 👉 this is where the magic happens, we compute batches on demand

	var result field.ElemSlice
	currentCol := s.currentCol

	if s.currentServer == s.numServers-1 {
		result = *s.lastCol

		for k := 0; k < s.numCols; k++ {
			for i := 0; i < s.colSize; i++ {

				index := (k * s.colSize) + i

				a := result.Get(index)
				a.Neg(&a)

				if currentCol == s.iy {
					a.Add(&a, &s.state.a[i])
				}

				result.Set(index, a)
			}

			currentCol++
		}

		s.currentServer++

		return result
	}

	result = field.NewElemSlice(s.colSize * s.numCols)
	result.SetRandom(rand.Reader)

	for k := 0; k < s.numCols; k++ {
		for i := 0; i < s.colSize; i++ {

			index := (k * s.colSize) + i

			sum := s.lastCol.Get(index)

			a := result.Get(index)
			sum.Add(&sum, &a)

			s.lastCol.Set(index, sum)
		}

		currentCol++
	}

	s.currentServer++

	return result
}

// 👉 first version of the iterator, which was less memory efficient because we
// had to keep batches for all servers at once in memory.

// func newQueryIterator(numCols, numServers, colSize, iy int, rnd io.Reader,
// 	state *state) *QueryIterator {

// 	return &QueryIterator{
// 		currentCol: 0,
// 		numCols:    numCols,
// 		numServers: numServers,
// 		colSize:    colSize,
// 		rnd:        rnd,
// 		iy:         iy,
// 		state:      state,
// 	}
// }

// QueryIterator is an iterator over all the columns in the database. Each
// element it this iterator contains all the queries for each server.
// type QueryIterator struct {
// 	currentCol int
// 	numCols    int
// 	numServers int
// 	colSize    int
// 	rnd        io.Reader

// 	iy    int
// 	state *state
// }

// HasNext ...
// func (q *QueryIterator) HasNext() bool {
// 	return q.currentCol < q.numCols
// }

// GetNext ...
// func (q *QueryIterator) GetNext() *field.ElemSliceIterator {
// 	result := make([]field.ElemSlice, q.numServers)

// 	for i := range result {
// 		result[i] = field.NewElemSlice(q.colSize)
// 		result[i].SetRandom(q.rnd)
// 	}

// 	for i := 0; i < q.colSize; i++ {
// 		sum := field.Zero()

// 		for k := 0; k < q.numServers-1; k++ {
// 			a := result[k].Get(i)
// 			sum.Add(&sum, &a)
// 		}

// 		sum.Neg(&sum)

// 		if q.currentCol == q.iy {
// 			sum.Add(&sum, &q.state.a[i])
// 		}

// 		result[q.numServers-1].Set(i, sum)
// 	}

// 	q.currentCol++

// 	return field.NewElemSliceIterator(result)
// }
