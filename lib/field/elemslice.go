package field

import (
	"encoding/binary"
	"io"
)

const esize = 8 * 2

// ElemSlice allows to use a slice of elements with a common underlying byte
// slice. It allows to "marshall" the slice with zero memory allocation.
type ElemSlice struct {
	n    int
	data []byte
}

// NewElemSlice created a new element slice.
func NewElemSlice(n int) ElemSlice {
	return ElemSlice{
		n:    n,
		data: make([]byte, n*esize),
	}
}

// NewElemSliceFromBytes creates a new element slice from an give byte slice.
func NewElemSliceFromBytes(data []byte) ElemSlice {
	return ElemSlice{
		n:    len(data) / esize,
		data: data,
	}
}

// Set sets the ith elements
func (e ElemSlice) Set(i int, el Element) {
	binary.LittleEndian.PutUint64(e.data[i*esize:i*esize+8], el[0])
	binary.LittleEndian.PutUint64(e.data[i*esize+8:i*esize+8+8], el[1])
}

// SetRandom fill the element slice with random values.
func (e ElemSlice) SetRandom(rnd io.Reader) error {
	_, err := io.ReadFull(rnd, e.data)
	if err != nil {
		return err
	}
	// TODO: SetFixedLength useful ..?

	return nil
}

// Get returns the ith element. Note that is returns a copy.s
func (e ElemSlice) Get(i int) Element {
	return Element{
		binary.LittleEndian.Uint64(e.data[i*esize : i*esize+8]),
		binary.LittleEndian.Uint64(e.data[i*esize+8 : i*esize+8+8]),
	}
}

// Range returns a range, equivalent to elements[begin:end]
func (e ElemSlice) Range(begin, end int) ElemSlice {
	if end == -1 {
		end = len(e.data) / esize
	}

	return ElemSlice{
		n:    end - begin,
		data: e.data[begin*esize : end*esize],
	}
}

// Bytes return the underlying bytes. Be careful with what you do with it as
// this is not a copy!
func (e ElemSlice) Bytes() []byte {
	return e.data
}

// 👉 not used anymore, left because some code still uses it

// ElemSliceIterator defines an iterator of elements.
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

func (e *ElemSliceIterator) GetNext() ElemSlice {
	res := e.data[e.current]
	e.current++

	return res
}
