package simulation

import "fmt"

type BlockResult struct {
	Query       float64
	Answer0     float64
	Answer1     float64
	Reconstruct float64
}

type DBResult struct {
	BlockResults []*BlockResult
	Total        float64
}

type ExperimentResult struct {
	ExperimentResult []*DBResult
}

func main() {
	fmt.Println("vim-go")
}
