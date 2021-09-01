package main

type Query uint8

const (
	QueryKeyId Query = iota
	QueryCreationTime
	QueryPubKeyAlgo
)
