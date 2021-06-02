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

func NewElemSliceFromBytes(data []byte) ElemSlice {
	return ElemSlice{
		n:    len(data) / esize,
		data: data,
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

func NewElemSliceGetter(c BytesChunks) ElemSliceGetter {
	return ElemSliceGetter{
		chunks: c,
	}
}

type ElemSliceGetter struct {
	chunks BytesChunks
}

func (e ElemSliceGetter) Get(i int) Element {
	return Element{
		binary.LittleEndian.Uint64(e.chunks.Range(i*esize, i*esize+8)),
		binary.LittleEndian.Uint64(e.chunks.Range(i*esize+8, i*esize+8+8)),
	}
}

func (e ElemSliceGetter) Range(begin, end int) ElemSliceGetter {
	if begin != 0 || end != -1 {
		panic("no implemented")
	}

	return e
}

func NewBytesChunks(data [][]byte, chunkSize int) BytesChunks {
	return BytesChunks{
		data:      data,
		chunkSize: chunkSize,
	}
}

type BytesChunks struct {
	data [][]byte
	// it assumes every chunk have the same size, except the last one
	chunkSize int
}

func (b BytesChunks) Get(index int) byte {
	chunkI := index / b.chunkSize
	i := index % chunkI
	return b.data[chunkI][i]
}

func (b BytesChunks) Range(begin, end int) []byte {
	beginChunkI := begin / b.chunkSize
	endChunkI := end / b.chunkSize

	beginI := begin % b.chunkSize
	endI := end % b.chunkSize

	if beginChunkI == endChunkI {
		return b.data[beginChunkI][beginI:endI]
	}

	res := make([]byte, 0, end-begin)

	// b.data:  [...][...][...][...]
	// res:       [.........]

	// rest of the first chunk
	res = append(res, b.data[beginChunkI][beginI:]...)

	// all intermediary chunks
	for i := beginChunkI + 1; i < endChunkI; i++ {
		res = append(res, b.data[i]...)
	}

	// rest of the last chunk
	res = append(res, b.data[endChunkI][:endI]...)

	return res
}
