package database

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudflare/circl/group"
	mmap "github.com/edsrzf/mmap-go"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/utils"
	"go.etcd.io/bbolt"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
)

var DefaultChunkSize = 1e7

const infoDbKey = "info"

func NewDB(info Info) (*DB, error) {
	n := info.BlockSize * info.NumColumns * info.NumRows
	if info.BlockSize == constants.SingleBitBlockLength {
		n = info.NumColumns * info.NumRows
	}

	return &DB{
		Info:     info,
		inMemory: make([]field.Element, n),
	}, nil
}

type DB struct {
	Info
	inMemory []field.Element

	// if true, the GetEntry uses the mmap, other the inMemory
	useMmap bool
	mmap    mmap.MMap
}

// String returns a string representation of a db. Elements are hashed.
func (d *DB) String() string {
	out := new(strings.Builder)

	h := sha256.New()
	n := d.Info.NumColumns * d.Info.NumRows
	b := make([]byte, 8)
	for i := 0; i < n; i++ {
		e := d.GetEntry(i)
		binary.LittleEndian.PutUint64(b, e[0])
		h.Write(b)
		binary.LittleEndian.PutUint64(b, e[1])
		h.Write(b)
	}
	els := h.Sum(nil)

	h = sha256.New()
	for _, el := range d.inMemory {
		b := el.Bytes()
		h.Write(b[:])
	}
	inMem := h.Sum(nil)

	h = sha256.New()
	h.Write(d.mmap)
	mmap := h.Sum(nil)

	fmt.Fprintf(out, "{DB: Info={%v} els={%x} inMemory={%x} useMmap=%v mmap={%x} }",
		d.Info, els, inMem, d.useMmap, mmap)

	return out.String()
}

func (d *DB) setEntry(i int, el field.Element) {
	d.inMemory[i] = el
}

func (d *DB) SizeGiB() float64 {
	return float64(len(d.mmap)) / (1024 * 1024 * 1024)
}

type saveInfo struct {
	Info Info
	// the list of chunks, with start/end indexes for each chunk
	Chunks [][2]int
}

// LoadInMemory loads the content of mmap in memory.
func (d *DB) LoadInMemory() {
	log.Println("loading db in memory")

	// Already in memory
	if !d.useMmap {
		log.Println("db already in memory, nothing to do")
		return
	}

	n := len(d.mmap) / (8 * 2)
	d.inMemory = make([]field.Element, n)

	for i := 0; i < n; i++ {
		d.inMemory[i] = d.GetEntry(i)
	}

	err := d.mmap.Unmap()
	if err != nil {
		log.Fatalf("failed to unmap memory: %v", err)
	}

	d.useMmap = false
}

func (d *DB) SaveDBFileSingle(root string) error {
	infoPath := filepath.Join(root, "info")

	chunkSize := 1e7

	infoFile, err := os.Create(infoPath)
	if err != nil {
		return xerrors.Errorf("failed to create info file: %v", err)
	}

	enc := gob.NewEncoder(infoFile)

	err = enc.Encode(&d.Info)
	if err != nil {
		infoFile.Close()
		return xerrors.Errorf("failed to encode info: %v", err)
	}

	infoFile.Close()

	log.Println("info file saved")

	outFile, err := os.Create(filepath.Join(root, "data"))
	if err != nil {
		return xerrors.Errorf("failed to create data file: %v", err)
	}

	// result := make([]byte, 8*2*len(d.inMemory))

	for i := 0; i < len(d.inMemory); i += int(chunkSize) {
		n := int(chunkSize)

		if i+int(chunkSize) >= len(d.inMemory) {
			n = len(d.inMemory) - i
		}

		log.Println("saving chunk", i, i+n)

		result := make([]byte, n*8*2)

		for k := 0; k < n; k++ {
			binary.LittleEndian.PutUint64(result[k*8*2:k*8*2+8], d.inMemory[k+i][0])
			binary.LittleEndian.PutUint64(result[k*8*2+8:k*8*2+8+8], d.inMemory[k+i][1])
		}

		_, err = outFile.Write(result)
		if err != nil {
			return xerrors.Errorf("failed to write bytes: %v", err)
		}

		outFile.Sync()
	}

	outFile.Close()

	return nil
}

func (d *DB) SaveDB(path string, bucket string) error {
	chunkSize := DefaultChunkSize

	db, err := bbolt.Open(path, 0666, nil)
	if err != nil {
		return xerrors.Errorf("failed to open db: %v", err)
	}

	defer db.Close()

	err = db.Update(func(t *bbolt.Tx) error {
		_, err := t.CreateBucket([]byte(bucket))
		if err != nil {
			return xerrors.Errorf("failed to create bucket: %v", err)
		}

		return nil
	})

	if err != nil {
		return xerrors.Errorf("failed to create bucket: %v", err)
	}

	saveInfo := saveInfo{
		Info:   d.Info,
		Chunks: make([][2]int, 0),
	}

	n := d.Info.BlockSize * d.Info.NumColumns * d.Info.NumRows

	err = db.Update(func(t *bbolt.Tx) error {
		for i := 0; i < n; i += int(chunkSize) {
			key := make([]byte, 8)
			binary.LittleEndian.PutUint64(key, uint64(i))

			var chunk []field.Element
			if i+int(chunkSize) >= n {
				chunk = d.inMemory[i:]
				log.Println("saving last chunk")
			} else {
				chunk = d.inMemory[i : i+int(chunkSize)]
			}

			buf := new(bytes.Buffer)
			enc := gob.NewEncoder(buf)

			err := enc.Encode(chunk)
			if err != nil {
				return xerrors.Errorf("failed to encode chunk: %v", err)
			}

			log.Println("saving chunk", i, i+len(chunk))
			saveInfo.Chunks = append(saveInfo.Chunks, [2]int{i, i + len(chunk)})

			err = t.Bucket([]byte(bucket)).Put(key, buf.Bytes())
			if err != nil {
				return xerrors.Errorf("failed to put chunk: %v", err)
			}

		}

		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)

		err := enc.Encode(&saveInfo)
		if err != nil {
			return xerrors.Errorf("failed to encode info: %v", err)
		}

		err = t.Bucket([]byte(bucket)).Put([]byte(infoDbKey), buf.Bytes())
		if err != nil {
			return xerrors.Errorf("failed to put info: %v", err)
		}

		return nil
	})

	if err != nil {
		return xerrors.Errorf("failed to save chunks: %v", err)
	}

	return nil
}

func LoadMMapDB(path string) (*DB, error) {
	infoFile, err := os.OpenFile(filepath.Join(path, "info"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, xerrors.Errorf("failed to open info: %v", err)
	}

	dec := gob.NewDecoder(infoFile)
	info := Info{}

	err = dec.Decode(&info)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode info: %v", err)
	}

	// https://github.com/golang/go/issues/4609
	if info.Auth == nil {
		info.Auth = &Auth{}
	}
	if info.Merkle == nil {
		info.Merkle = &Merkle{}
	}
	if info.DataEmbedding == nil {
		info.DataEmbedding = &DataEmbedding{}
	}
	if info.LatParams == nil {
		info.LatParams = &bfv.Parameters{}
	}

	f, err := os.OpenFile(filepath.Join(path, "data"), os.O_RDONLY, 0644)
	if err != nil {
		return nil, xerrors.Errorf("failed to read file: %v", err)
	}

	mmap, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		return nil, xerrors.Errorf("failed to map: %v", err)
	}

	db := DB{
		Info:    info,
		useMmap: true,
		mmap:    mmap,
	}

	return &db, nil
}

func LoadDB(path, bucket string) (*DB, error) {
	db, err := bbolt.Open(path, 0666, nil)
	if err != nil {
		return nil, xerrors.Errorf("failed to open db: %v", err)
	}

	defer db.Close()

	var elements []field.Element
	var info Info
	saveInfo := saveInfo{}
	var n int

	err = db.View(func(t *bbolt.Tx) error {

		res := t.Bucket([]byte(bucket)).Get([]byte(infoDbKey))
		buf := bytes.NewBuffer(res)
		dec := gob.NewDecoder(buf)

		err := dec.Decode(&saveInfo)
		if err != nil {
			return xerrors.Errorf("failed to decode info: %v", err)
		}

		info = saveInfo.Info
		n = info.BlockSize * info.NumColumns * info.NumRows

		return nil
	})

	if err != nil {
		return nil, xerrors.Errorf("failed to read info: %v", err)
	}

	key := make([]byte, 8)

	err = db.View(func(t *bbolt.Tx) error {
		elements = make([]field.Element, n)

		for _, i := range saveInfo.Chunks {
			start, end := i[0], i[1]

			chunk := make([]field.Element, end-start)

			binary.LittleEndian.PutUint64(key, uint64(start))

			res := t.Bucket([]byte(bucket)).Get(key)
			buf := bytes.NewBuffer(res)

			dec := gob.NewDecoder(buf)
			err = dec.Decode(&chunk)
			if err != nil {
				return xerrors.Errorf("failed to decode chunk: %v", err)
			}

			log.Println("loading", start, start+len(chunk))
			copy(elements[start:start+len(chunk)], chunk)
		}

		return nil
	})

	if err != nil {
		return nil, xerrors.Errorf("failed to read chunks: %v", err)
	}

	result := DB{
		inMemory: elements,
		Info:     info,
	}

	return &result, nil
}

func (d *DB) GetEntry(i int) field.Element {
	if !d.useMmap {
		return d.inMemory[i]
	}

	memIndex := i * 8 * 2

	return field.Element{
		binary.LittleEndian.Uint64(d.mmap[memIndex : memIndex+8]),
		binary.LittleEndian.Uint64(d.mmap[memIndex+8 : memIndex+16]),
	}
}

// Range returns a view on the db, scoped by a begin-end index.
func (d *DB) Range(begin, end int) ElementRange {
	if end == -1 {
		end = len(d.mmap) / (8 * 2)
	}

	return ElementRange{
		db:    d,
		begin: begin,
		len:   end - begin,
	}
}

// ElementRange defines a range on the db, this allows to have a sort of
// slice[a:b] function. It doesn't load new element in memory but uses the same
// db.
type ElementRange struct {
	db    *DB
	begin int
	len   int
}

// Get return the ith element relative to the range, from 0:(end-start).
func (f *ElementRange) Get(i int) field.Element {
	return f.db.GetEntry(i + f.begin)
}

// Range creates a range on a range:
// db.mmap:                [....................]
// db.Range:                ....[...........]...
// range.Range:             ........[....]......
func (f *ElementRange) Range(begin, end int) ElementRange {
	return ElementRange{
		db:    f.db,
		begin: begin + f.begin,
		len:   end - begin,
	}
}

// Len returns the remaining number of elements
func (f *ElementRange) Len() int {
	return f.len
}

type Info struct {
	NumRows    int
	NumColumns int
	BlockSize  int

	// PIR type: classical, merkle, signature
	PIRType string

	*Auth
	*Merkle
	*DataEmbedding

	//Lattice parameters for the single-server data retrieval
	LatParams *bfv.Parameters
}

// Authentication information for the single-server setting
type Auth struct {
	// The global digest that is a hash of all the row digests. Public.
	Digest []byte
	// One digest per row, authenticating all the elements in that row.
	SubDigests []byte
	// ECC group and hash algorithm used for digest computation and PIR itself
	Group group.Group
	Hash  crypto.Hash
	// Due to lack of the size functions in the lib API, we store it in the db info
	ElementSize int
	ScalarSize  int
}

// Data embedding info
type DataEmbedding struct {
	IDLength  int
	KeyLength int
}

// The info needed for the Merkle-tree based approach
type Merkle struct {
	Root     []byte
	ProofLen int
}

func CreateZeroMultiBitDB(numRows, numColumns, blockSize int) (*DB, error) {
	info := Info{NumColumns: numColumns,
		NumRows:   numRows,
		BlockSize: blockSize,
	}

	db, err := NewDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	n := numRows * numColumns * blockSize
	for i := 0; i < n; i++ {
		db.setEntry(i, field.Zero())
	}

	return db, nil
}

func CreateRandomMultiBitDB(rnd io.Reader, dbLen, numRows, blockLen int) (*DB, error) {
	numColumns := dbLen / (8 * field.Bytes * numRows * blockLen)
	// handle very small db
	if numColumns == 0 {
		numColumns = 1
	}

	info := Info{
		NumColumns: numColumns,
		NumRows:    numRows,
		BlockSize:  blockLen,
	}

	n := numRows * numColumns * blockLen

	bytesLength := n*field.Bytes + 1
	bytes := make([]byte, bytesLength)

	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, xerrors.Errorf("failed to read random bytes: %v", err)
	}

	db, err := NewDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	for i := 0; i < n; i++ {
		var buf [16]byte
		copy(buf[:], bytes[i*field.Bytes:(1+i)*field.Bytes])
		element := &field.Element{}
		element.SetFixedLengthBytes(buf)

		db.setEntry(i, *element)
	}

	return db, nil
}

func CreateRandomSingleBitDB(rnd io.Reader, dbLen, numRows int) (*DB, error) {
	numColumns := dbLen / numRows

	// by convention a block size of 0 indicates the single-bit scheme
	info := Info{
		NumColumns: numColumns,
		NumRows:    numRows,
		BlockSize:  1,
	}

	db, err := NewDB(info)
	if err != nil {
		return nil, xerrors.Errorf("failed to create db: %v", err)
	}

	bytes := make([]byte, dbLen)
	if _, err := io.ReadFull(rnd, bytes[:]); err != nil {
		return nil, xerrors.Errorf("failed to read random bytes: %v", err)
	}

	for i := 0; i < dbLen; i++ {
		element := field.Element{}

		if bytes[i]>>7 == 1 {
			element.SetOne()
		} else {
			element.SetZero()
		}

		db.setEntry(i, element)
	}

	return db, nil
}

// HashToIndex hashes the given id to an index for a database of the given
// length
func HashToIndex(id string, length int) int {
	hash := blake2b.Sum256([]byte(id))
	return int(binary.BigEndian.Uint64(hash[:]) % uint64(length))
}

func CalculateNumRowsAndColumns(numBlocks int, matrix bool) (numRows, numColumns int) {
	if matrix {
		utils.IncreaseToNextSquare(&numBlocks)
		numColumns = int(math.Sqrt(float64(numBlocks)))
		numRows = numColumns
	} else {
		numColumns = numBlocks
		numRows = 1
	}
	return
}
