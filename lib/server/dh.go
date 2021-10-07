package server

import (
	"math"
	"runtime"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/database"
)

// A DH server for the single-server DL-based tag retrieval
type DH struct {
	db *database.Elliptic
}

func NewDH(db *database.Elliptic) *DH {
	return &DH{db: db}
}

func (s *DH) AnswerBytes(q []byte) ([]byte, error) {
	query, err := database.UnmarshalGroupElements(q, s.db.Group, s.db.ElementSize)
	if err != nil {
		return nil, err
	}

	NGoRoutines := runtime.NumCPU()
	// make sure that we do not need up with routines processing 0 elements
	if NGoRoutines > s.db.NumRows {
		NGoRoutines = s.db.NumRows
	}
	rowsPerRoutine := int(math.Ceil(float64(s.db.NumRows) / float64(NGoRoutines)))
	replies := make([]chan []group.Element, NGoRoutines)
	var begin, end int
	for i := 0; i < NGoRoutines; i++ {
		begin, end = i*rowsPerRoutine, (i+1)*rowsPerRoutine
		// make the last routine take all the left-over (from division) rows
		if end > s.db.NumRows {
			end = s.db.NumRows
		}
		replyChan := make(chan []group.Element)
		replies[i] = replyChan
		go s.processRows(begin, end, query, replyChan)
	}

	answer := make([]group.Element, 0, s.db.NumRows)
	for i, reply := range replies {
		chunk := <-reply
		answer = append(answer, chunk...)
		close(replies[i])
	}

	// Encode the answer into binary
	encoded, err := database.MarshalGroupElements(answer, s.db.ElementSize)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func (s *DH) processRows(begin, end int, input []group.Element, replyTo chan<- []group.Element) {
	// one product per row
	prods := make([]group.Element, end-begin)
	for i := begin; i < end; i++ {
		prods[i-begin] = s.db.Group.Identity()
		for j := 0; j < s.db.NumColumns; j++ {
			if s.db.Entries[i*s.db.NumColumns+j] == 1 {
				// add query element to the product if
				// the corresponding database bit is 1
				prods[i-begin].Add(prods[i-begin], input[j])
			}
		}
	}
	replyTo <- prods
}
