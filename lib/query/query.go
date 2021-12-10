package query

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
	"strconv"
	"time"

	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/fss"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

// Target defines the target of the query
type Target uint8

const (
	// UserId corresponds to the email
	UserId Target = iota

	CreationTime

	// RSA, ED25519, ...
	PubKeyAlgo
)

// ClientFSS is used by the client to prepare an FSS
type ClientFSS struct {
	*Info
	Input []bool
}

// FSS is what is sent to the server, one by server
type FSS struct {
	*Info
	FssKey fss.FssKeyEq2P
}

// Info defines the query function
type Info struct {
	// Target is on what the query is to be executed. An email (id), or the
	// creation time.
	Target Target

	// select the substring on the id (email)
	FromStart, FromEnd int // start and end of the target

	And     bool
	Targets []Target

	// to perform AVG query
	Avg bool

	// to perform SUM query
	// TODO: not implemented yet, but implicitely used in AVG
	Sum bool
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
	v := &ClientFSS{}
	err := dec.Decode(v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (i *Info) ToEmailClientFSS(in string) *ClientFSS {
	id, _ := i.IdForEmail(in)
	return &ClientFSS{
		Info:  i,
		Input: id,
	}
}

func (i *Info) ToPKAClientFSS(in string) *ClientFSS {
	var pka packet.PublicKeyAlgorithm
	switch in {
	case "RSA":
		pka = packet.PubKeyAlgoRSA
	case "ElGamal":
		pka = packet.PubKeyAlgoElGamal
	case "DSA":
		pka = packet.PubKeyAlgoDSA
	case "ECDH":
		pka = packet.PubKeyAlgoECDH
	case "ECDSA":
		pka = packet.PubKeyAlgoECDSA
	}
	id := i.IdForPubKeyAlgo(pka)
	return &ClientFSS{
		Info:  i,
		Input: id,
	}
}

func (i *Info) ToCreationTimeClientFSS(in string) *ClientFSS {
	year, err := strconv.Atoi(in)
	if err != nil {
		log.Fatal(err)
	}
	match := time.Date(year, 0, 0, 0, 0, 0, 0, time.UTC)
	id, err := i.IdForCreationTime(match)
	if err != nil {
		panic(err)
	}
	return &ClientFSS{
		Info:  i,
		Input: id,
	}
}

// TODO: hardcoded for the moment, FIX
func (i *Info) ToAndClientFSS(in string) *ClientFSS {
	idYear, err := i.IdForYearCreationTime(time.Date(2019, 0, 0, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	idOrganization, _ := i.IdForEmail(in)

	// combine
	combined := append(idYear, idOrganization...)
	return &ClientFSS{
		Info:  i,
		Input: combined,
	}
}

func (q *FSS) IdForEmail(email string) ([]bool, bool) {
	return q.Info.IdForEmail(email)
}

func (q *FSS) IdForPubKeyAlgo(pka packet.PublicKeyAlgorithm) []bool {
	return q.Info.IdForPubKeyAlgo(pka)
}

func (q *FSS) IdForCreationTime(t time.Time) ([]bool, error) {
	return q.Info.IdForCreationTime(t)
}

func (q *FSS) IdForYearCreationTime(t time.Time) ([]bool, error) {
	return q.Info.IdForYearCreationTime(t)
}

func (i *Info) IdForEmail(email string) ([]bool, bool) {
	var id []bool
	if i.FromStart != 0 {
		if i.FromStart > len(email) {
			return nil, false
		}
		id = utils.ByteToBits([]byte(email[:i.FromStart]))
	} else if i.FromEnd != 0 {
		if i.FromEnd > len(email) {
			return nil, false
		}
		id = utils.ByteToBits([]byte(email[len(email)-i.FromEnd:]))
	} else {
		h := blake2b.Sum256([]byte(email))
		id = utils.ByteToBits(h[:16])
	}

	return id, true
}

func (i *Info) IdForPubKeyAlgo(pka packet.PublicKeyAlgorithm) []bool {
	return utils.ByteToBits([]byte{uint8(pka)})
}

func (i *Info) IdForCreationTime(t time.Time) ([]bool, error) {
	binaryMatch, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return utils.ByteToBits(binaryMatch), nil
}

func (i *Info) IdForYearCreationTime(t time.Time) ([]bool, error) {
	y := uint32(t.Year())
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, y)
	return utils.ByteToBits(b), nil
}
