package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"runtime"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/fss"
)

type FSS struct {
	db    *database.DB
	cores int

	serverNum byte
	fss       *fss.Fss
}

// use variadic argument for cores to achieve backward compatibility
func NewFSS(db *database.DB, serverNum byte, prfKeys [][]byte, cores ...int) *FSS {
	numCores := runtime.NumCPU()
	if len(cores) > 0 {
		numCores = cores[0]
	}

	return &FSS{
		db:        db,
		cores:     numCores,
		serverNum: serverNum,
		fss:       fss.ServerInitialize(prfKeys, field.Bits),
	}

}

func (s *FSS) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *FSS) AnswerBytes(q []byte) ([]byte, error) {
	// decode query
	buf := bytes.NewBuffer(q)
	dec := gob.NewDecoder(buf)
	var query fss.FssKeyEq2P
	if err := dec.Decode(&query); err != nil {
		return nil, err
	}

	// get answer
	a := s.Answer(query)

	// encode answer
	buf.Reset()
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(a); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *FSS) Answer(key fss.FssKeyEq2P) []uint32 {
	idLength := constants.IdentifierLength
	numIdentifiers := s.db.NumColumns

	q := make([]uint32, (s.db.BlockSize+1)*numIdentifiers)
	tmp := make([]uint32, s.db.BlockSize+1)
	for i := 0; i < numIdentifiers; i++ {
		id := binary.BigEndian.Uint32(s.db.Identifiers[i*idLength : (i+1)*idLength])
		s.fss.EvaluatePF(s.serverNum, key, id, tmp)
		copy(q[i*(s.db.BlockSize+1):(i+1)*(s.db.BlockSize+1)], tmp)
	}

	return answer(q, s.db, s.cores)
}
