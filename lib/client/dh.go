package client

import (
	"bytes"
	"errors"
	"io"
	"log"
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
// the single-server scheme
func NewDH(rnd io.Reader, info *database.Info) *DH {
	// check that row digests hash to the global one
	hasher := info.Hash.New()
	hasher.Write(info.SubDigests)
	if !bytes.Equal(hasher.Sum(nil), info.Digest) {
		panic("row digests and the global digest in the info do not match")
	}
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
		replyChan := make(chan []group.Element, end-begin)
		replies[i] = replyChan
		go generateBlindedElements(begin, end, r, c.dbInfo, replyChan)
	}

	// Combine the generated chunks from all the routines.
	// We wait for each routines in the initial order so
	// it is ok to simply append the results one after another.
	query := make([]group.Element, 0, c.dbInfo.NumColumns*c.dbInfo.BlockSize)
	for i, reply := range replies {
		chunk := <-reply
		query = append(query, chunk...)
		close(replies[i])
	}

	// Add the additional blinding t to the the retrieval index.
	// See Construction 9 of the paper.
	st.ht = database.CommitScalarToIndex(t, uint64(st.iy), g)
	query[st.iy].Add(query[st.iy], st.ht)
	c.state = st

	encodedQuery, err := database.MarshalGroupElements(query, c.dbInfo.ElementSize)
	if err != nil {
		return nil, err
	}

	return encodedQuery, nil
}

func (c *DH) ReconstructBytes(a []byte) (interface{}, error) {
	g := c.dbInfo.Group
	digSize := c.dbInfo.ElementSize
	rneg := g.NewScalar().Neg(c.state.r)
	// get the tags of all the rows
	answer, err := database.UnmarshalGroupElements(a, c.dbInfo.Group, c.dbInfo.ElementSize)
	if err != nil {
		return nil, err
	}
	m := g.Identity()
	var res byte
	for i := 0; i < c.dbInfo.NumRows; i++ {
		// get the row digest and raise it to a power r
		d := g.NewElement()
		err = d.UnmarshalBinary(c.dbInfo.SubDigests[i*digSize : (i+1)*digSize])
		if err != nil {
			return nil, err
		}
		d.Mul(d, rneg)
		m.Add(d, answer[i])
		if !m.IsIdentity() && !m.IsEqual(c.state.ht) {
			return nil, errors.New("reject")
		}
		if i == c.state.ix {
			switch {
			case m.IsIdentity():
				res = 0
			case m.IsEqual(c.state.ht):
				res = 1
			default:
				log.Printf("something wrong, accepted %v\n", m)
			}
		}
	}

	return res, nil
}

// Hash indices to group elements and multiply by the blinding scalar
func generateBlindedElements(begin, end int, blinding group.Scalar, info *database.Info, replyTo chan<- []group.Element) {
	elements := make([]group.Element, 0, (end-begin)*info.BlockSize)
	for j := begin; j < end; j++ {
		elements = append(elements, database.CommitScalarToIndex(blinding, uint64(j), info.Group))
	}
	replyTo <- elements
}
