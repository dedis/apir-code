package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"math/bits"
	"math/rand"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cloudflare/circl/group"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/field"
	"github.com/si-co/vpir-code/lib/monitor"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/server"
	"github.com/si-co/vpir-code/lib/utils"
)

const generalConfigFile = "simul.toml"

type generalParam struct {
	DBBitLengths   []int
	BitsToRetrieve int
	Repetitions    int
}

type individualParam struct {
	Name           string
	Primitive      string
	NumServers     []int
	NumRows        int
	BlockLength    int
	ElementBitSize int
	InputSizes     []int // FSS input sizes in bytes
}

type Simulation struct {
	generalParam
	individualParam
}

func main() {
	// seed non-cryptographic randomness
	rand.Seed(time.Now().UnixNano())

	// create results directory if not presenc
	folderPath := "results"
	if _, err := os.Stat(folderPath); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write mem profile to file")
	indivConfigFile := flag.String("config", "", "config file for simulation")
	flag.Parse()

	// CPU profiling
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	// make sure cfg file is specified
	if *indivConfigFile == "" {
		panic("simulation's config file not provided")
	}
	log.Printf("config file %s", *indivConfigFile)

	// load simulation's config files
	s, err := loadSimulationConfigs(generalConfigFile, *indivConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	// check simulation
	if !s.validSimulation() {
		log.Fatal("invalid simulation")
	}

	log.Printf("running simulation %#v\n", s)
	// initialize experiment
	experiment := &Experiment{Results: make(map[int][]*Chunk, 0)}
	var experimentv *Experiment
	if s.Primitive[:3] == "fss" {
		experimentv = &Experiment{Results: make(map[int][]*Chunk, 0)}
	}

	// range over all the DB lengths specified in the general simulation config
dbSizesLoop:
	for _, dl := range s.DBBitLengths {
		// compute database data
		dbLen := dl
		blockLen := s.BlockLength
		elemBitSize := s.ElementBitSize
		nRows := s.NumRows

		// Find the total number of blocks in the db
		numBlocks := dl
		if s.Primitive[:3] != "cmp" {
			numBlocks = dbLen / (elemBitSize * blockLen)
		}
		// matrix db
		if nRows != 1 {
			utils.IncreaseToNextSquare(&numBlocks)
			nRows = int(math.Sqrt(float64(numBlocks)))
		}

		// setup db, this is the same for DPF or IT
		dbPRG := utils.RandomPRG()
		db := new(database.DB)
		dbBytes := new(database.Bytes)
		dbRing := new(database.Ring)
		dbElliptic := new(database.Elliptic)
		dbLWE := new(database.LWE)
		switch s.Primitive[:3] {
		case "pir":
			if s.Primitive[len(s.Primitive)-6:] == "merkle" {
				dbBytes = database.CreateRandomMerkle(dbPRG, dbLen, nRows, blockLen)
			} else {
				dbBytes = database.CreateRandomBytes(dbPRG, dbLen, nRows, blockLen)
			}
		case "fss":
			// num of identifiers in the random FSS database
			numIdenfitiers := 100000
			db, err = database.CreateRandomKeysDB(dbPRG, numIdenfitiers)
			if err != nil {
				panic(err)
			}
		case "cmp":
			if s.Primitive == "cmp-pir" {
				log.Printf("Generating lattice db of size %d\n", dbLen)
				dbRing = database.CreateRandomRingDB(dbPRG, dbLen, true)
			} else if s.Primitive == "cmp-vpir" {
				log.Printf("Generating elliptic db of size %d\n", dbLen)
				dbElliptic = database.CreateRandomEllipticWithDigest(dbPRG, dbLen, group.P256, true)
			} else if s.Primitive == "cmp-vpir-lwe" {
				log.Printf("Generating LWE db of size %d\n", dbLen)
				dbLWE = database.CreateRandomBinaryLWEWithLength(dbPRG, dbLen)
			}
		}

		// GC after DB creation
		runtime.GC()
		time.Sleep(3)

		// run experiment
		var results []*Chunk
		switch s.Primitive {
		case "pir-classic", "pir-merkle":
			if len(s.NumServers) == 0 {
				log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits",
					s.Primitive, dbLen)
				numServers := 2                                   // basic case with two servers
				blockSize := dbBytes.BlockSize - dbBytes.ProofLen // ProofLen = 0 for PIR
				results = pirIT(dbBytes, numServers, blockSize, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
			} else {
				// if NumServers is specified, we loop only through the number of servers
				for _, numServers := range s.NumServers {
					log.Printf("retrieving blocks with primitive %s from DB with dbLen = %d bits from %d servers",
						s.Primitive, dbLen, numServers)
					blockSize := dbBytes.BlockSize - dbBytes.ProofLen // ProofLen = 0 for PIR
					results = pirIT(dbBytes, numServers, blockSize, s.ElementBitSize, s.BitsToRetrieve, s.Repetitions)
					experiment.Results[numServers] = results
				}

				time.Sleep(3)
				runtime.GC()
				time.Sleep(3)

				// Skip the rest of the loop
				break dbSizesLoop
			}
		case "fss":
			// In FSS, we iterate over input sizes instead of db sizes
			for _, inputSize := range s.InputSizes {
				stringToSearch := utils.Ranstring(inputSize)
				// Non-verifiable FSS
				results = fssPIR(db, inputSize, stringToSearch, s.Repetitions)
				log.Printf("retrieving with non-verifiable FSS with input size of %d bytes\n", inputSize)
				experiment.Results[inputSize] = results
				// Authenticated FSS
				results = fssVPIR(db, inputSize, stringToSearch, s.Repetitions)
				log.Printf("retrieving with verifiable FSS with the input size of %d bytes\n", inputSize)
				experimentv.Results[inputSize] = results

				time.Sleep(3)
				runtime.GC()
				time.Sleep(3)
			}
			// Skip the rest of the loop
			break dbSizesLoop
		case "cmp-pir":
			log.Printf("db info: %#v", dbRing.Info)
			results = pirLattice(dbRing, s.Repetitions)
		case "cmp-vpir":
			log.Printf("db info: %#v", dbElliptic.Info)
			results = pirElliptic(dbElliptic, s.Repetitions)
		case "cmp-vpir-lwe":
			log.Printf("db info: %#v", dbLWE.Info)
			results = pirLWE(dbLWE, s.Repetitions)
		default:
			log.Fatal("unknown primitive type")
		}
		experiment.Results[dbLen] = results

		// GC at the end of the iteration
		runtime.GC()
	}

	// print results
	res, err := json.Marshal(experiment)
	if err != nil {
		panic(err)
	}
	fileName := s.Name + ".json"
	if err = ioutil.WriteFile(path.Join("results", fileName), res, 0644); err != nil {
		panic(err)
	}

	if s.Primitive[:3] == "fss" {
		resv, err := json.Marshal(experimentv)
		if err != nil {
			panic(err)
		}
		fileName := "auth" + s.Name + ".json"
		if err = ioutil.WriteFile(path.Join("results", fileName), resv, 0644); err != nil {
			panic(err)
		}
	}

	// mem profiling
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
	log.Println("simulation terminated successfully")
}

func fssVPIR(db *database.DB, inputSize int, stringToSearch string, nRepeat int) []*Chunk {
	c := client.NewPredicateAPIR(utils.RandomPRG(), &db.Info)
	ss := []*server.PredicateAPIR{server.NewPredicateAPIR(db, 0), server.NewPredicateAPIR(db, 1)}

	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	in := utils.ByteToBits([]byte(stringToSearch))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: inputSize},
		Input: in,
	}
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(1)
		results[j].CPU[0] = initBlock(len(ss))
		results[j].Bandwidth[0] = initBlock(len(ss))

		m.Reset()
		queries := c.Query(q, 2)
		results[j].CPU[0].Query = m.RecordAndReset()
		for r := range queries {
			results[j].Bandwidth[0].Query += fssQueryByteLength(queries[r])
		}

		// get servers answers
		answers := make([][]uint32, len(ss))
		for k := range ss {
			m.Reset()
			answers[k] = ss[k].Answer(queries[k])
			results[j].CPU[0].Answers[k] = m.RecordAndReset()
			results[j].Bandwidth[0].Answers[k] = fieldVectorByteLength(answers[k])
		}

		m.Reset()
		_, err := c.Reconstruct(answers)
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		if err != nil {
			log.Fatal(err)
		}
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func fssPIR(db *database.DB, inputSize int, stringToSearch string, nRepeat int) []*Chunk {
	c := client.NewPredicatePIR(utils.RandomPRG(), &db.Info)
	ss := []*server.PredicatePIR{server.NewPredicatePIR(db, 0), server.NewPredicatePIR(db, 1)}

	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	in := utils.ByteToBits([]byte(stringToSearch))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: inputSize},
		Input: in,
	}
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(1)
		results[j].CPU[0] = initBlock(len(ss))
		results[j].Bandwidth[0] = initBlock(len(ss))

		m.Reset()
		queries := c.Query(q, 2)
		results[j].CPU[0].Query = m.RecordAndReset()
		for r := range queries {
			results[j].Bandwidth[0].Query += fssQueryByteLength(queries[r])
		}

		// get servers answers
		answers := make([][]uint32, len(ss))
		for k := range ss {
			m.Reset()
			answers[k] = ss[k].Answer(queries[k])
			results[j].CPU[0].Answers[k] = m.RecordAndReset()
			results[j].Bandwidth[0].Answers[k] = float64(bits.UintSize / 8)
		}

		m.Reset()
		_, err := c.Reconstruct(answers)
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		if err != nil {
			log.Fatal(err)
		}
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func pirIT(db *database.Bytes, numServers, blockSize, elemBitSize, numBitsToRetrieve, nRepeat int) []*Chunk {
	prg := utils.RandomPRG()
	c := client.NewPIR(prg, &db.Info)
	ss := makePIRServers(db, numServers)
	numTotalBlocks := db.NumRows * db.NumColumns
	numRetrieveBlocks := bitsToBlocks(blockSize, elemBitSize, numBitsToRetrieve)

	// create main monitor for CPU time
	m := monitor.NewMonitor()

	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	var startIndex int
	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrieveBlocks)

		// pick a random block index to start the retrieval
		startIndex = rand.Intn(numTotalBlocks - numRetrieveBlocks)
		for i := 0; i < numRetrieveBlocks; i++ {
			results[j].CPU[i] = initBlock(len(ss))
			results[j].Bandwidth[i] = initBlock(len(ss))

			m.Reset()
			queries := c.Query(startIndex+i, len(ss))
			results[j].CPU[i].Query = m.RecordAndReset()
			for r := range queries {
				results[j].Bandwidth[i].Query += float64(len(queries[r]))
			}

			// get servers answers
			answers := make([][]byte, len(ss))
			for k := range ss {
				m.Reset()
				answers[k] = ss[k].Answer(queries[k])
				results[j].CPU[i].Answers[k] = m.RecordAndReset()
				results[j].Bandwidth[i].Answers[k] = float64(len(answers[k]))
			}

			m.Reset()
			_, err := c.Reconstruct(answers)
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			results[j].Bandwidth[i].Reconstruct = 0
			if err != nil {
				log.Fatal(err)
			}
		}

		// GC after each repetition
		runtime.GC()

		// sleep after every iteration
		time.Sleep(2 * time.Second)
	}

	return results
}

func pirLattice(db *database.Ring, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	c := client.NewLattice(&db.Info)
	s := server.NewLattice(db)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index := rand.Intn(db.NumRows * db.NumColumns)
		for i := 0; i < numRetrievedBlocks; i++ {
			results[j].CPU[i] = initBlock(1)
			results[j].Bandwidth[i] = initBlock(1)

			m.Reset()
			query, err := c.QueryBytes(index + i)
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Query = m.RecordAndReset()
			results[j].Bandwidth[i].Query = float64(len(query))

			answer, err := s.AnswerBytes(query)
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Answers[0] = m.RecordAndReset()
			results[j].Bandwidth[i].Answers[0] = float64(len(answer))

			_, err = c.ReconstructBytes(answer)
			if err != nil {
				log.Fatal(err)
			}
			results[j].CPU[i].Reconstruct = m.RecordAndReset()
			results[j].Bandwidth[i].Reconstruct = 0
		}

		// GC after each repetition
		runtime.GC()
		time.Sleep(2)
	}
	return results
}

func pirLWE(db *database.LWE, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	p := utils.ParamsWithDatabaseSize(db.Info.NumRows, db.Info.NumColumns)
	c := client.NewLWE(utils.RandomPRG(), &db.Info, p)
	s := server.NewLWE(db)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index := rand.Intn(db.NumRows * db.NumColumns)
		results[j].CPU[0] = initBlock(1)
		results[j].Bandwidth[0] = initBlock(1)

		m.Reset()
		query, err := c.QueryBytes(index)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Query = m.RecordAndReset()
		results[j].Bandwidth[0].Query += float64(len(query))

		// get server's answer
		answer, err := s.AnswerBytes(query)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Answers[0] = m.RecordAndReset()
		results[j].Bandwidth[0].Answers[0] = float64(len(answer))

		_, err = c.ReconstructBytes(answer)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()
		time.Sleep(2)
	}

	return results
}

func pirElliptic(db *database.Elliptic, nRepeat int) []*Chunk {
	numRetrievedBlocks := 1
	// create main monitor for CPU time
	m := monitor.NewMonitor()
	// run the experiment nRepeat times
	results := make([]*Chunk, nRepeat)

	prg := utils.RandomPRG()
	c := client.NewDH(prg, &db.Info)
	s := server.NewDH(db)

	for j := 0; j < nRepeat; j++ {
		log.Printf("start repetition %d out of %d", j+1, nRepeat)
		results[j] = initChunk(numRetrievedBlocks)
		// pick a random block index to start the retrieval
		index := rand.Intn(db.NumRows * db.NumColumns)
		results[j].CPU[0] = initBlock(1)
		results[j].Bandwidth[0] = initBlock(1)

		m.Reset()
		query, err := c.QueryBytes(index)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Query = m.RecordAndReset()
		results[j].Bandwidth[0].Query += float64(len(query))

		// get server's answer
		answer, err := s.AnswerBytes(query)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Answers[0] = m.RecordAndReset()
		results[j].Bandwidth[0].Answers[0] = float64(len(answer))

		_, err = c.ReconstructBytes(answer)
		if err != nil {
			log.Fatal(err)
		}
		results[j].CPU[0].Reconstruct = m.RecordAndReset()
		results[j].Bandwidth[0].Reconstruct = 0

		// GC after each repetition
		runtime.GC()
		time.Sleep(2)
	}

	return results
}

// Converts number of bits to retrieve into the number of db blocks
func bitsToBlocks(blockSize, elemSize, numBits int) int {
	return int(math.Ceil(float64(numBits) / float64(blockSize*elemSize)))
}

func makePIRServers(db *database.Bytes, numServers int) []*server.PIR {
	servers := make([]*server.PIR, numServers)
	for i := range servers {
		servers[i] = server.NewPIR(db)
	}
	return servers
}

func fssQueryByteLength(q *query.FSS) float64 {
	totalLen := 0

	// Count the bytes of FssKey
	totalLen += len(q.FssKey.SInit)
	totalLen += 1 // q.FssKey.TInit
	totalLen += len(q.FssKey.FinalCW) * field.Bytes
	for i := range q.FssKey.CW {
		totalLen += len(q.FssKey.CW[i])
	}

	// Count the bytes of Info
	totalLen += infoSizeByte(q.Info)

	return float64(totalLen)
}

func infoSizeByte(q *query.Info) int {
	totalLen := 0
	// Count the bytes of Info
	// q.Target and q.Targets are uint8 and []uint8,
	// respectively
	totalLen += len(q.Targets) + 1 // q.Target
	// The size of int is platform dependent
	totalLen += bits.UintSize / 8 //q.FromStart
	totalLen += bits.UintSize / 8 // q.FromEnd
	// And is bool
	totalLen += 1

	return totalLen
}

func fieldVectorByteLength(vec []uint32) float64 {
	return float64(len(vec) * field.Bytes)
}

func initChunk(numRetrieveBlocks int) *Chunk {
	return &Chunk{
		CPU:       make([]*Block, numRetrieveBlocks),
		Bandwidth: make([]*Block, numRetrieveBlocks),
	}
}

func initBlock(numAnswers int) *Block {
	return &Block{
		Query:       0,
		Answers:     make([]float64, numAnswers),
		Reconstruct: 0,
	}
}

func loadSimulationConfigs(genFile, indFile string) (*Simulation, error) {
	var err error
	genConfig := new(generalParam)
	_, err = toml.DecodeFile(genFile, genConfig)
	if err != nil {
		return nil, err
	}
	indConfig := new(individualParam)
	_, err = toml.DecodeFile(indFile, indConfig)
	if err != nil {
		return nil, err
	}
	return &Simulation{generalParam: *genConfig, individualParam: *indConfig}, nil
}

func (s *Simulation) validSimulation() bool {
	return s.Primitive == "pir-classic" ||
		s.Primitive == "pir-merkle" ||
		s.Primitive == "fss" ||
		s.Primitive == "cmp-pir" ||
		s.Primitive == "cmp-vpir" ||
		s.Primitive == "cmp-vpir-lwe"
}
