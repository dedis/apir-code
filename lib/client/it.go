package client

import (
	"encoding/binary"
	"io"
	"log"

	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
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
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

func (c *IT) QueryBytes(index, numServers int) ([][]byte, error) {
	// get reconstruction
	queries := c.Query(index, numServers)

	// encode all the queries in bytes
	out := make([][]byte, len(queries))
	for i := range queries {
		queryBuf := make([]byte, len(queries[i])*8*2)
		for k := 0; k < len(queries[i]); k++ {
			binary.LittleEndian.PutUint64(queryBuf[k*8*2:k*8*2+8], queries[i][k][0])
			binary.LittleEndian.PutUint64(queryBuf[k*8*2+8:k*8*2+8+8], queries[i][k][1])
		}
		out[i] = queryBuf
	}
	//for i, q := range queries {
	//buf := new(bytes.Buffer)
	//enc := gob.NewEncoder(buf)
	//if err := enc.Encode(q); err != nil {
	//return nil, err
	//}
	//out[i] = buf.Bytes()
	//}

	return out, nil
}

// Query performs a client query for the given database index to numServers
// servers. This function performs both vector and rebalanced query depending
// on the client initialization.
func (c *IT) Query(index, numServers int) [][]field.Element {
	if invalidQueryInputsIT(index, numServers) {
		log.Fatal("invalid query inputs")
	}

	var err error
	c.state, err = generateClientState(index, c.rnd, c.dbInfo)
	if err != nil {
		log.Fatal(err)
	}

	vectors, err := c.secretShare(numServers)
	if err != nil {
		log.Fatal(err)
	}
	return vectors
}

// ReconstructBytes returns []field.Element
func (c *IT) ReconstructBytes(a [][]byte) (interface{}, error) {
	//answer, err := decodeAnswer(a)
	//if err != nil {
	//return nil, err
	//}
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
func (c *IT) secretShare(numServers int) ([][]field.Element, error) {
	// get block length
	blockLen := len(c.state.a)
	// Number of field elements in the whole vector
	vectorLen := c.dbInfo.NumColumns * blockLen

	// create query vectors for all the servers F^(1+b)
	vectors := make([][]field.Element, numServers)
	for k := range vectors {
		vectors[k] = make([]field.Element, vectorLen)
	}

	// Get random elements for all numServers-1 vectors
	rand, err := field.RandomVector(c.rnd, (numServers-1)*vectorLen)
	if err != nil {
		return nil, err
	}
	// perform additive secret sharing
	var colStart, colEnd int
	for j := 0; j < c.dbInfo.NumColumns; j++ {
		colStart = j * blockLen
		colEnd = (j + 1) * blockLen
		// Assign k - 1 random vectors of length dbLength containing
		// elements in F^(1+b)
		for k := 0; k < numServers-1; k++ {
			copy(vectors[k][colStart:colEnd], rand[k*vectorLen+colStart:k*vectorLen+colEnd])
		}

		// we should perform component-wise additive secret sharing
		for b := colStart; b < colEnd; b++ {
			sum := field.Zero()
			for k := 0; k < numServers-1; k++ {
				sum.Add(&sum, &vectors[k][b])
			}
			vectors[numServers-1][b].Set(&sum)
			vectors[numServers-1][b].Neg(&vectors[numServers-1][b])
			// set alpha vector at the block we want to retrieve
			if j == c.state.iy {
				vectors[numServers-1][b].Add(&vectors[numServers-1][b], &c.state.a[b-j*blockLen])
			}
		}
	}

	return vectors, nil
}
