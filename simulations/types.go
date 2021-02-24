package main

type Block struct {
	Query       float64
	Answers     []float64
	Reconstruct float64
}

type Chunk struct {
	CPU            []*Block
	Bandwidth      []*Block
}

type Experiment struct {
	Results map[int][]*Chunk
}

const (
	oneMB = 1048576 * 8
	oneKB = 1024 * 8
)
