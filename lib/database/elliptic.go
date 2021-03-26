package database

import (
	"crypto"
	"encoding/binary"
	"github.com/si-co/vpir-code/lib/utils"
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
	elementSize, scalarSize := getElementAndScalarSizes(g)
	preSquareNumBlocks := dbLen / (8 * scalarSize * blockLen)
	numRows, numColumns := CalculateNumRowsAndColumns(preSquareNumBlocks, rebalanced)

	h := crypto.BLAKE2b_256
	// fill out the db with scalars and
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
	hasher.Write(digests)

	return &Elliptic{Entries: entries,
		Digests: digests,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
			Auth: &Auth{
				Digest:      hasher.Sum(nil),
				Group:       g,
				Hash:        h,
				ElementSize: elementSize,
				ScalarSize:  scalarSize,
			},
		},
	}
}

func CreateEllipticWithDigestFromData(data []byte, info *Info) *Elliptic {
	var err error
	g := group.P256
	elementSize, scalarSize := getElementAndScalarSizes(g)
	numRows, numColumns := info.NumRows, info.NumColumns
	blockLen := len(data) / (numRows * numColumns * scalarSize)

	h := crypto.BLAKE2b_256
	// fill out the db with scalars and
	// compute the db digest.
	entries := make([]group.Scalar, numRows*numColumns*blockLen)
	digests := make([]byte, 0, numRows*h.Size())
	for i := 0; i < numRows; i++ {
		d := g.Identity()
		for j := 0; j < numColumns; j++ {
			for l := 0; l < blockLen; l++ {
				// get scalar from the data
				s := g.NewScalar()
				err = s.UnmarshalBinary(data[(i*numColumns*blockLen+j*blockLen+l)*scalarSize : (i*numColumns*blockLen+j*blockLen+l+1)*scalarSize])
				if err != nil {
					log.Fatal(err)
				}
				entries[i*numColumns*blockLen+j*blockLen+l] = s
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
	hasher.Write(digests)

	return &Elliptic{Entries: entries,
		Digests: digests,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: blockLen,
			Auth: &Auth{
				Digest:      hasher.Sum(nil),
				Group:       g,
				Hash:        h,
				ElementSize: elementSize,
				ScalarSize:  scalarSize,
			},
		},
	}
}

func getElementAndScalarSizes(g group.Group) (int, int) {
	// Obtaining the scalar and element sizes for the group
	rnd := utils.RandomPRG()
	rndScalar, _ := g.RandomScalar(rnd).MarshalBinary()
	rndElement, _ := g.RandomElement(rnd).MarshalBinaryCompress()
	return len(rndElement), len(rndScalar)
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
