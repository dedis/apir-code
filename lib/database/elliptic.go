package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"log"

	"github.com/cloudflare/circl/group"
)

type Elliptic struct {
	Entries []group.Scalar
	// One digest per row, authenticating all the elements in that row.
	Digests []byte
	Info
}

// blockLen must be the number of scalars in a block
func CreateRandomEllipticWithDigest(rnd io.Reader, g group.Group, dbLen, blockLen int, rebalanced bool) *Elliptic {
	// Obtaining the scalar and element sizes for the group
	rndScalar, _ := g.RandomScalar(rnd).MarshalBinary()
	rndElement, _ := g.RandomElement(rnd).MarshalBinaryCompress()
	scalarSize := len(rndScalar)
	elementSize := len(rndElement)
	h := crypto.BLAKE2b_256

	preSquareNumBlocks := dbLen / (8 * scalarSize * blockLen)
	numRows, numColumns := CalculateNumRowsAndColumns(preSquareNumBlocks, rebalanced)

	// fill out the db with random scalars and
	// compute the db digest.
	entries := make([]group.Scalar, numRows*numColumns*blockLen)
	digests := make([]byte, 0, numRows*h.Size())
	for i := 0; i < numRows; i++ {
		d := g.Identity()
		for j := 0; j < numColumns; j++ {
			for l := 0; l < blockLen; l++ {
				// sample and fill in a random scalar
				entries[i*numColumns*blockLen+j*blockLen+l] = g.RandomScalar(rnd)
				d.Add(d, CommitScalarToIndex(entries[i*numColumns*blockLen+j*blockLen+l], uint64(j), uint64(l), g))
			}
		}
		tmp, err := d.MarshalBinaryCompress()
		if err != nil {
			log.Fatal(err)
		}
		digests = append(digests, tmp...)
	}
	// global digest
	hasher := h.New()
	return &Elliptic{Entries: entries,
		Digests: digests,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
			Auth: &Auth{
				Digest:      hasher.Sum(digests),
				Group:       g,
				Hash:        h,
				ElementSize: elementSize,
				ScalarSize:  scalarSize,
			},
		},
	}
}

// Take the indices (j, l) and hash them to get a group element
func HashIndexToGroup(j, l uint64, g group.Group) group.Element {
	// hash the concatenation of row and block indices to a group element
	index := make([]byte, 16)
	binary.LittleEndian.PutUint64(index[:8], j)
	binary.LittleEndian.PutUint64(index[8:], l)
	return g.HashToElement(index, nil)
}

// Raise the group element obtained iva index hashing to the scalar
func CommitScalarToIndex(x group.Scalar, j, l uint64, g group.Group) group.Element {
	H := HashIndexToGroup(j, l, g)
	// multiply the hash by the db entry as a scalar
	return g.NewElement().Mul(H, x)
}

// Marshal a slice of group elements
func MarshalGroupElements(q []group.Element, marshalledLen int) ([]byte, error) {
	encoded := make([]byte, 0, marshalledLen*len(q))
	for _, el := range q {
		tmp, err := el.MarshalBinaryCompress()
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, tmp...)
	}
	return encoded, nil
}

// Unmarshal a slice of group elements
func UnmarshalGroupElements(q []byte, g group.Group, elemSize int) ([]group.Element, error) {
	var err error
	decoded := make([]group.Element, 0, len(q)/elemSize)
	for i := 0; i < len(q); i += elemSize {
		elem := g.NewElement()
		err = elem.UnmarshalBinary(q[i : i+elemSize])
		if err != nil {
			return nil, err
		}
		decoded = append(decoded, elem)
	}
	return decoded, nil
}
