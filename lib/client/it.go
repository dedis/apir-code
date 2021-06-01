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
		rnd:    rand.Reader,
		dbInfo: info,
		state:  nil,
	}
}

func (c *IT) QueryBytes(index, numServers int) ([][]byte, error) {
	panic("not implemented")
}

func (c *IT) QueryBytesNew(index, numServers int) (*QueryIterator, error) {
	// get reconstruction
	queries := c.Query(index, numServers)

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
func (c *IT) Query(index, numServers int) *QueryIterator {
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

	iter := newQueryIterator(c.dbInfo.NumColumns, numServers, len(c.state.a),
		c.state.iy, c.rnd, c.state)

	return iter
}

// ReconstructBytes returns []field.Element
func (c *IT) ReconstructBytes(a [][]byte) (interface{}, error) {
	// answer, err := decodeAnswer(a)
	// if err != nil {
	// 	return nil, err
	// }

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

func newQueryIterator(numCols, numServers, colSize, iy int, rnd io.Reader,
	state *state) *QueryIterator {

	return &QueryIterator{
		currentCol: 0,
		numCols:    numCols,
		numServers: numServers,
		colSize:    colSize,
		rnd:        rnd,
		iy:         iy,
		state:      state,
	}
}

// QueryIterator ...
type QueryIterator struct {
	currentCol int
	numCols    int
	numServers int
	colSize    int
	rnd        io.Reader

	iy    int
	state *state
}

// HasNext ...
func (q *QueryIterator) HasNext() bool {
	return q.currentCol < q.numCols
}

// Get ...
func (q *QueryIterator) Get() *field.ElemSliceIterator {
	result := make([]field.ElemSlice, q.numServers)

	for i := range result {
		result[i] = field.NewElemSlice(q.colSize)
		result[i].SetRandom(q.rnd)
	}

	for i := 0; i < q.colSize; i++ {
		sum := field.Zero()

		for k := 0; k < q.numServers-1; k++ {
			a := result[k].Get(i)
			sum.Add(&sum, &a)
		}

		sum.Neg(&sum)

		if q.currentCol == q.iy {
			sum.Add(&sum, &q.state.a[i])
		}

		result[q.numServers-1].Set(i, sum)
	}

	q.currentCol++

	return field.NewElemSliceIterator(result)
}
