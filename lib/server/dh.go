package server

import (
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
	rowsPerRoutine := s.db.NumRows / NGoRoutines
	replies := make([]chan []group.Element, NGoRoutines)
	var begin, end int
	for i := 0; i < NGoRoutines; i++ {
		begin, end = i*rowsPerRoutine, (i+1)*rowsPerRoutine
		// make the last routine take all the left-over (from division) rows
		if i == NGoRoutines-1 {
			end = s.db.NumRows
		}
		replyChan := make(chan []group.Element, rowsPerRoutine)
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
	// Appending row digests to the answer
	encoded = append(encoded, s.db.Digests...)

	return encoded, nil
}

func (s *DH) processRows(begin, end int, input []group.Element, replyTo chan<- []group.Element) {
	// one tag per row
	tags := make([]group.Element, end-begin)
	for i := begin; i < end; i++ {
		tags[i-begin] = s.db.Group.Identity()
		for j := 0; j < s.db.NumColumns; j++ {
			for l := 0; l < s.db.BlockSize; l++ {
				// multiply an element of the query by the corresponding scalar from the db
				tmp := s.db.Group.NewElement()
				tmp.Mul(input[j*s.db.BlockSize+l], s.db.Entries[i*s.db.NumColumns*s.db.BlockSize+j*s.db.BlockSize+l])
				// sum up the result into the row tag
				tags[i-begin].Add(tags[i-begin], tmp)
			}
		}
	}
	replyTo <- tags
}
