package matrix

type MatrixBytes struct {
	rows int
	cols int
	data []byte
}

func NewBytes(r int, c int) *MatrixBytes {
	return &MatrixBytes{
		rows: r,
		cols: c,
		data: make([]byte, r*c),
	}
}

// func (m *MatrixBytes) SetData(d []byte) {
// 	m.Data = d
// }

// func (m *MatrixBytes) Get(r int, c int) byte {
// 	return (m.Data[(m.cols*r+c)/8] >> (m.cols*r + c) % 8) & 1
// }

func (m *MatrixBytes) SetData(i int, v byte) {
	m.data[i] = v
}

func (m *MatrixBytes) Get(r int, c int) byte {
	return m.data[(m.cols*r + c)]
}

func (m *MatrixBytes) Len() int {
	return len(m.data)
}
