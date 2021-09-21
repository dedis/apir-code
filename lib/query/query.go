package query

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/authfss"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
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

type AuthFSS struct {
	AdditionalInformationFSS
	FssKey authfss.FssKeyEq2P
}

type FSS struct {
	AdditionalInformationFSS
	FssKey fss.FssKeyEq2P
}

type AdditionalInformationFSS struct {
	Target             Target
	FromStart, FromEnd int // start and end of the target

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

func (q AuthFSS) IdForEmail(email string) ([]bool, bool) {
	var id []bool
	if q.FromStart != 0 {
		if q.FromStart > len(email) {
			return nil, false
		}
		id = utils.ByteToBits([]byte(email[:q.FromStart]))
	} else if q.FromEnd != 0 {
		if q.FromEnd > len(email) {
			return nil, false
		}
		id = utils.ByteToBits([]byte(email[len(email)-q.FromEnd:]))
	} else {
		h := blake2b.Sum256([]byte(email))
		id = utils.ByteToBits(h[:16])
	}

	return id, true
}

func (q AuthFSS) IdForPubKeyAlgo(pka packet.PublicKeyAlgorithm) []bool {
	return utils.ByteToBits([]byte{uint8(pka)})
}

func (q AuthFSS) IdForCreationTime(t time.Time) ([]bool, error) {
	binaryMatch, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return utils.ByteToBits(binaryMatch), nil
}
