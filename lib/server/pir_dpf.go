package server

import (
	"math/bits"

	"github.com/lukechampine/fastxor"
	"github.com/si-co/vpir-code/lib/database"

	"github.com/dimakogan/dpf-go/dpf"
)

type PIRdpf struct {
	db        *database.Bytes
	serverNum byte
}

func NewPIRdpf(db *database.Bytes, serverNum byte) *PIRdpf {
	return &PIRdpf{db: db,
		serverNum: serverNum,
	}
}

func (s *PIRdpf) DBInfo() *database.Info {
	return &s.db.Info
}

func (s *PIRdpf) AnswerBytes(q []byte) ([]byte, error) {
	panic("not yet implemented")
	return nil, nil
}

func (s *PIRdpf) Answer(key dpf.DPFkey) []byte {
	q := dpf.EvalFull(key, uint64(bits.Len(uint(s.db.NumColumns)-1)))

	m := make([]byte, s.db.NumRows*s.db.BlockSize)
	// we have to traverse column by column
	for i := 0; i < s.db.NumRows; i++ {
		sum := make([]byte, s.db.BlockSize)
		for j := 0; j < s.db.NumColumns; j++ {
			if (q[j/8]>>(j%8))&1 == byte(1) {
				fastxor.Bytes(sum, sum, s.db.Entries[i][j*s.db.BlockSize:j*s.db.BlockSize+s.db.BlockSize])
			}
		}
		copy(m[i*(s.db.BlockSize):(i+1)*(s.db.BlockSize)], sum)
	}

	return m
}
