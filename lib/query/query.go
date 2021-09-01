package query

import "github.com/si-co/vpir-code/lib/fss"

type Type uint8

const (
	KeyId Type = iota
	CreationTime
	PubKeyAlgo
)

type FSS struct {
	QueryType Type
	FssKey    fss.FssKeyEq2P
}
