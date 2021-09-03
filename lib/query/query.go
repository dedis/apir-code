package query

import "github.com/si-co/vpir-code/lib/fss"

type Target uint8

const (
	UserId Target = iota
	CreationTime
	PubKeyAlgo
	Key
)

type FSS struct {
	Target             Target
	FromStart, FromEnd int // start and end of the target
	FssKey             fss.FssKeyEq2P
}
