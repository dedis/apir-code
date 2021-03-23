package server

import (
	"runtime"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/utils"
)

type Single struct {
	db *database.Elliptic
}

func NewSingle(db *database.Elliptic) *Single {
	return &Single{db: db}
}

func (s *Single) AnswerBytes(q []byte) ([]byte, error) {
	query, err := database.UnmarshalGroupElements(q, s.db.Group, s.db.ElementSize)
	if err != nil {
		return nil, err
	}

	NGoRoutines := runtime.NumCPU()
	rowsPerRoutine := utils.DivideAndRoundUpToMultiple(s.db.NumRows, NGoRoutines, 1)
	replies := make([]chan []group.Element, NGoRoutines)
	var begin, end int
	for i := 0; i < NGoRoutines; i++ {
		begin, end = i*rowsPerRoutine, (i+1)*rowsPerRoutine
		if end >= s.db.NumRows {
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

func (s *Single) processRows(begin, end int, input []group.Element, replyTo chan<- []group.Element) {
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
