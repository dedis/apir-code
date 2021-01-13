package client

import (
	"errors"

	cst "github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
)

// Client represents the client instance in both the IT and C models
type Client interface {
	Query()
	Reconstruct()
}

// General containts the elements needed by the clients of all schemes
type General struct {
	DBLength int
}

func reconstruct(answers [][]field.Element, blockSize int, index int, alpha field.Element, a []field.Element) ([]field.Element, error) {
	answersLen := len(answers[0])
	sum := make([]field.Element, answersLen)

	// sum answers as vectors in F(2^128)^(1+b)
	for i := 0; i < answersLen; i++ {
		sum[i] = field.Zero()
		for s := range answers {
			sum[i].Add(&sum[i], &answers[s][i])
		}
	}

	if blockSize == cst.SingleBitBlockLength {
		switch {
		case sum[index].Equal(&alpha):
			return []field.Element{cst.One}, nil
		case sum[index].Equal(&cst.Zero):
			return []field.Element{cst.Zero}, nil
		default:
			return nil, errors.New("reject")
		}
	}

	tag := sum[len(sum)-1]
	messages := sum[:len(sum)-1]

	// compute reconstructed tag
	reconstructedTag := field.Zero()
	for i := 0; i < len(messages); i++ {
		var prod field.Element
		prod.Mul(&a[i], &messages[i])
		reconstructedTag.Add(&reconstructedTag, &prod)
	}

	if !tag.Equal(&reconstructedTag) {
		return nil, errors.New("REJECT")
	}

	return messages, nil
}

// return true if the query inputs are invalid
func invalidQueryInputs(index, blockSize, numServers int) bool {
	return (index < 0 || blockSize < 1 || index > cst.DBLength) && numServers < 1
}
