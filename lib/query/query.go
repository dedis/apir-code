package query

import (
	"bytes"
	"encoding/gob"

	"github.com/si-co/vpir-code/lib/fss"
)

type Target uint8

const (
	UserId Target = iota
	CreationTime
	PubKeyAlgo
)

// TODO: refactor into a single type FSS and then differenciate in Input and FssKey?
// TODO: this needs refactoring
type ClientFSS struct {
	Target             Target
	FromStart, FromEnd int // start and end of the target
	Input              []bool

	And     bool
	Targets []Target
}

type FSS struct {
	Target             Target
	FromStart, FromEnd int // start and end of the target
	FssKey             fss.FssKeyEq2P

	And     bool
	Targets []Target
}

func (q *ClientFSS) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(q); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeClientFSS(in []byte) (*ClientFSS, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(in))
	var v *ClientFSS
	err := dec.Decode(v)
	if err != nil {
		return nil, err
	}

	return v, nil
}
