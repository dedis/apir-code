package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/si-co/vpir-code/lib/field"
	"math/bits"
	"runtime"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
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
		fss:       fss.ServerInitialize(prfKeys, uint(bits.Len(uint(db.Info.NumColumns)-1))),
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
	numIdentifiers := len(s.db.Identifiers) / idLength

	sum := uint32(0)
	q := make([]uint32, (s.db.BlockSize+1)*numIdentifiers)
	for i := 0; i < numIdentifiers; i++ {
		tmp := make([]uint32, s.db.BlockSize+1)
		id := binary.BigEndian.Uint32(s.db.Identifiers[i*idLength : (i+1)*idLength])
		s.fss.EvaluatePF(s.serverNum, key, uint(id), tmp)
		sum = (sum + tmp[0]) % field.ModP
		copy(q[i*(s.db.BlockSize+1):(i+1)*(s.db.BlockSize+1)], tmp)
	}
	fmt.Printf("%d - %d\n", s.serverNum, sum)
	return answer(q, s.db, s.cores)
}
