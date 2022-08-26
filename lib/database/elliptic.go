package database

import (
	"crypto"
	"encoding/binary"
	"io"
	"log"
	"math"
	"runtime"

	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/utils"
)

type Elliptic struct {
	Entries []byte
	Info
}

func CreateRandomEllipticWithDigest(rnd io.Reader, dbLen int, g group.Group, rebalanced bool) *Elliptic {
	numRows, numColumns := CalculateNumRowsAndColumns(dbLen, rebalanced)
	// read random bytes for filling out the entries
	// For simplicity, we use the whole byte to store 0 or 1
	data := make([]byte, numRows*numColumns)
	if _, err := rnd.Read(data); err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(data); i++ {
		data[i] = data[i] & 1
	}
	NGoRoutines := runtime.NumCPU()
	// for small dbs use one core
	if dbLen <= 1048576 {
		NGoRoutines = 4
	}
	h := crypto.BLAKE2b_256
	rowsPerRoutine := int(math.Ceil(float64(numRows) / float64(NGoRoutines)))
	replies := make([]chan []byte, NGoRoutines)
	var begin, end int
	for i := 0; i < NGoRoutines; i++ {
		begin, end = i*rowsPerRoutine, (i+1)*rowsPerRoutine
		// make the last routine take all the left-over (from division) rows
		if end > numRows {
			end = numRows
		}
		replyChan := make(chan []byte, 1)
		replies[i] = replyChan
		go computeDigests(begin, end, data, numColumns, g, h, replyChan)
	}
	digests := make([]byte, 0, numRows*h.Size())
	for i, reply := range replies {
		chunk := <-reply
		digests = append(digests, chunk...)
		close(replies[i])
	}

	// global digest
	hasher := h.New()
	hasher.Write(digests)

	return &Elliptic{Entries: data,
		Info: Info{NumColumns: numColumns,
			NumRows:   numRows,
			BlockSize: 1,
			Auth: &Auth{
				Digest:      hasher.Sum(nil),
				SubDigests:  digests,
				Group:       g,
				Hash:        h,
				ElementSize: getGroupElementSize(g),
			},
		},
	}
}

func computeDigests(begin, end int, data []byte, rowLen int, g group.Group, h crypto.Hash, replyTo chan<- []byte) {
	digs := make([]byte, 0, (end-begin)*h.Size())
	for i := begin; i < end; i++ {
		d := g.Identity()
		for j := 0; j < rowLen; j++ {
			if data[i*rowLen+j] == 1 {
				d.Add(d, HashIndexToGroup(uint64(j), g))
			}
		}
		tmp, err := d.MarshalBinaryCompress()
		if err != nil {
			log.Fatal(err)
		}
		digs = append(digs, tmp...)
	}
	replyTo <- digs
}

// Take the indices (j, l) and hash them to get a group element
func HashIndexToGroup(j uint64, g group.Group) group.Element {
	// hash the concatenation of row and block indices to a group element
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, j)
	return g.HashToElement(index, nil)
}

// Raise the group element obtained via index hashing to the scalar
func CommitScalarToIndex(x group.Scalar, j uint64, g group.Group) group.Element {
	H := HashIndexToGroup(j, g)
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

func getGroupElementSize(g group.Group) int {
	// Obtaining the scalar and element sizes for the group
	rnd := utils.RandomPRG()
	rndElement, _ := g.RandomElement(rnd).MarshalBinaryCompress()
	return len(rndElement)
}
