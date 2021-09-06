package query

import "github.com/si-co/vpir-code/lib/fss"

type Target uint8

const (
	UserId Target = iota
	CreationTime
	PubKeyAlgo
	Key
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
