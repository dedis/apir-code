package field

import (
	"encoding/binary"
	"io"
)

const esize = 8 * 2

// ElemSlice ...
type ElemSlice struct {
	n    int
	data []byte
}

// NewElemSlice ...
func NewElemSlice(n int) ElemSlice {
	return ElemSlice{
		n:    n,
		data: make([]byte, n*esize),
	}
}

// Set ...
func (e ElemSlice) Set(i int, el Element) {
	binary.LittleEndian.PutUint64(e.data[i*esize:i*esize+8], el[0])
	binary.LittleEndian.PutUint64(e.data[i*esize+8:i*esize+8+8], el[1])
}

// SetRandom ...
func (e ElemSlice) SetRandom(rnd io.Reader) error {
	_, err := io.ReadFull(rnd, e.data)
	if err != nil {
		return err
	}
	// TODO: SetFixedLength useful ..?

	return nil
}

// Get ...
func (e ElemSlice) Get(i int) Element {
	return Element{
		binary.LittleEndian.Uint64(e.data[i*esize : i*esize+8]),
		binary.LittleEndian.Uint64(e.data[i*esize+8 : i*esize+8+8]),
	}
}

// Bytes return the underlying bytes. Be careful with what you do with it as
// this is not a copy!
func (e ElemSlice) Bytes() []byte {
	return e.data
}

type ElemSliceIterator struct {
	data    []ElemSlice
	current int
}

func NewElemSliceIterator(data []ElemSlice) *ElemSliceIterator {
	return &ElemSliceIterator{
		data:    data,
		current: 0,
	}
}

func (e *ElemSliceIterator) HasNext() bool {
	return e.current < len(e.data)
}

func (e *ElemSliceIterator) Get() ElemSlice {
	res := e.data[e.current]
	e.current++

	return res
}
