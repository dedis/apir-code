package client

import (
	"bytes"
	"errors"
	"io"
	"runtime"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/database"
)

// Single-server tag retrieval scheme
type DH struct {
	rnd    io.Reader
	dbInfo *database.Info
	state  *state
}

// NewDH returns an instance of a DH-based client for
// the single-server tag retrieval scheme
func NewDH(rnd io.Reader, info *database.Info) *DH {
	return &DH{
		rnd:    rnd,
		dbInfo: info,
		state:  nil,
	}
}

// QueryBytes takes as input the index of an entry in the database and returns
// the query for the server encoded in bytes.
func (c *DH) QueryBytes(index int) ([]byte, error) {
	g := c.dbInfo.Group

	// sample two random scalars
	r, t := g.RandomScalar(c.rnd), g.RandomScalar(c.rnd)

	// initialize state
	st := &state{}

	// compute the position in the db (vector or matrix)
	// if db is a vector, ix always equals 0
	st.ix = index / c.dbInfo.NumColumns
	st.iy = index % c.dbInfo.NumColumns
	st.r = r

	// multithreading
	NGoRoutines := runtime.NumCPU()
	columnsPerRoutine := c.dbInfo.NumColumns / NGoRoutines
	replies := make([]chan []group.Element, NGoRoutines)
	for i := 0; i < NGoRoutines; i++ {
		begin, end := i*columnsPerRoutine, (i+1)*columnsPerRoutine
		if i == NGoRoutines-1 {
			end = c.dbInfo.NumColumns
		}
		replyChan := make(chan []group.Element, columnsPerRoutine*c.dbInfo.BlockSize)
		replies[i] = replyChan
		go generateBlindedElements(begin, end, r, c.dbInfo, replyChan)
	}

	// Combine the generated chunks from all the routines.
	// We wait for each routines in the initial order so it is ok
	// to simply append the results one after another.
	query := make([]group.Element, 0, c.dbInfo.NumColumns*c.dbInfo.BlockSize)
	for i, reply := range replies {
		chunk := <-reply
		query = append(query, chunk...)
		close(replies[i])
	}

	// Add the additional blinding t to the block of the retrieval index.
	// See Construction 10 of the paper.
	T := copyScalar(t, g)
	hTs := make([]group.Element, c.dbInfo.BlockSize)
	for l := 0; l < c.dbInfo.BlockSize; l++ {
		hT := database.CommitScalarToIndex(T, uint64(st.iy), uint64(l), g)
		pos := st.iy*c.dbInfo.BlockSize + l
		query[pos].Add(query[pos], hT)
		hTs[l] = hT
		T.Mul(T, t)
	}
	st.Ht = hTs
	c.state = st

	encodedQuery, err := database.MarshalGroupElements(query, c.dbInfo.ElementSize)
	if err != nil {
		return nil, err
	}

	return encodedQuery, nil
}

func (c *DH) ReconstructBytes(a []byte, m [][]byte) (interface{}, error) {
	g := c.dbInfo.Group
	// get row digests from the end of the answer
	digSize := c.dbInfo.ElementSize
	digestsSize := c.dbInfo.NumRows * digSize
	digests := a[len(a)-digestsSize:]
	// check that row digests hash to the global one
	hasher := c.dbInfo.Hash.New()
	hasher.Write(digests)
	if !bytes.Equal(hasher.Sum(nil), c.dbInfo.Digest) {
		return nil, errors.New("received row digests and the global digests do not match")
	}

	// get the tags of all the rows
	answer, err := database.UnmarshalGroupElements(a[:len(a)-digestsSize], c.dbInfo.Group, c.dbInfo.ElementSize)
	if err != nil {
		return nil, err
	}
	scalarSize := c.dbInfo.ScalarSize
	ml := g.NewScalar()
	for i := 0; i < c.dbInfo.NumRows; i++ {
		// multiply all the block elements
		sum := g.Identity()
		for l := 0; l < c.dbInfo.BlockSize; l++ {
			h := g.NewElement()
			err = ml.UnmarshalBinary(m[i][l*scalarSize : (l+1)*scalarSize])
			if err != nil {
				return nil, err
			}
			h.Mul(c.state.Ht[l], ml)
			sum.Add(sum, h)
		}
		// get the row digest and raise it to a power r
		d := g.NewElement()
		err = d.UnmarshalBinary(digests[i*digSize : (i+1)*digSize])
		if err != nil {
			return nil, err
		}
		d.Mul(d, c.state.r)
		tau := g.NewElement().Add(d, sum)
		if !tau.IsEqual(answer[i]) {
			return nil, errors.New("the tag is incorrect")
		}
	}

	return nil, nil
}

// Hash indices to group elements and multiply by the blinding scalar
func generateBlindedElements(begin, end int, blinding group.Scalar, info *database.Info, replyTo chan<- []group.Element) {
	elements := make([]group.Element, 0, (end-begin)*info.BlockSize)
	for j := begin; j < end; j++ {
		for l := 0; l < info.BlockSize; l++ {
			elements = append(elements, database.CommitScalarToIndex(blinding, uint64(j), uint64(l), info.Group))
		}
	}
	replyTo <- elements
}

// A hack function (due to lack of lib API) to return a copy of a scalar
func copyScalar(scalar group.Scalar, g group.Group) group.Scalar {
	data, err := scalar.MarshalBinary()
	if err != nil {
		panic(err)
	}
	s := g.NewScalar()
	err = s.UnmarshalBinary(data)
	if err != nil {
		panic(err)
	}
	return s
}
