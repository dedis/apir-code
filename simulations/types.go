package main

type BlockResult struct {
	Query       float64
	Answer0     float64
	Answer1     float64
	Reconstruct float64
}

type DBResult struct {
	Results []*BlockResult
	Total   float64
}

type Experiment struct {
	Results map[int][]*DBResult
}

const (
	oneMB = 1048576 * 8
	oneKB = 1024 * 8
)
