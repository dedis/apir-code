package main

// Test suite for integrated VPIR.

import (
	"runtime"
	"testing"
	"time"

	"github.com/nikirill/go-crypto/openpgp/packet"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/crypto/blake2b"
)

const (
	oneB            = 8
	oneKB           = 1024 * oneB
	oneMB           = 1024 * oneKB
	testBlockLength = 64
	numIdentifiers  = 1000
)

var randomDB *database.DB

func initRandomDB() {
	rndrandomDB := utils.RandomPRG()
	var err error
	randomDB, err = database.CreateRandomKeysDB(rndrandomDB, numIdentifiers)
	if err != nil {
		panic(err)
	}

	// GC after DB creation
	runtime.GC()
}

func TestCountEntireEmail(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedEmailMatch(randomDB)

	retrieveComplex(t, randomDB, q, match, "TestCountEntireEmail")
}

func TestCountEntireEmailPIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedEmailMatch(randomDB)

	retrieveComplexPIR(t, randomDB, q, match, "TestCountEntireEmailPIR")
}

func TestCountStartsWithEmail(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedStartsWithMatch(db)

	retrieveComplex(t, randomDB, q, match, "TestCountStartsWithEmail")
}

func TestCountStartsWithEmailPIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedStartsWithMatch(db)

	retrieveComplexPIR(t, randomDB, q, match, "TestCountStartsWithEmailPIR")
}

func TestCountEndsWithEmail(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedEndsWithMatch(db)

	retrieveComplex(t, randomDB, q, match, "TestCountEndsWithEmail")
}

func TestCountEndsWithEmailPIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedEndsWithMatch(db)

	retrieveComplexPIR(t, randomDB, q, match, "TestCountEndsWithEmailPIR")
}

func TestCountPublicKeyAlgorithm(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedPkaMatch(db)

	retrieveComplex(t, randomDB, q, match, "TestCountPublicKeyAlgorithm")
}

func TestCountPublicKeyAlgorithmPIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedPkaMatch(db)

	retrieveComplexPIR(t, randomDB, q, match, "TestCountPublicKeyAlgorithmPIR")
}

func TestCountCreationTime(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedCreationTimeMatch(db)

	retrieveComplex(t, randomDB, q, match, "TestCreationDate")
}

func TestCountCreationTimePIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedCreationTimeMatch(db)

	retrieveComplexPIR(t, randomDB, q, match, "TestCreationDatePIR")
}

func TestCountAndQuery(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedAndQueryMatch(db)

	retrieveComplex(t, randomDB, q, match, "TestCountAndQuery")
}

func TestCountAndQueryPIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedAndQueryMatch(db)

	retrieveComplexPIR(t, randomDB, q, match, "TestCountAndQueryPIR")
}

func fixedAndQueryMatch(db *database.DB) (interface{}, *query.ClientFSS) {
	match := []interface{}{time.Date(2009, time.November, 0, 0, 0, 0, 0, time.UTC), packet.PubKeyAlgoRSA}

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].CreationTime = match[0].(time.Time)
		randomDB.KeysInfo[i].PubKeyAlgo = match[1].(packet.PublicKeyAlgorithm)
	}

	//matchByte:= append([]byte(match[0].(string)), byte(match[1].(packet.PublicKeyAlgorithm)))
	matchBytes, err := match[0].(time.Time).MarshalBinary()
	if err != nil {
		panic(err)
	}
	matchBytes = append(matchBytes, byte(match[1].(packet.PublicKeyAlgorithm)))
	in := utils.ByteToBits(matchBytes)
	q := &query.ClientFSS{
		Info: &query.Info{
			And:     true,
			Targets: []query.Target{query.PubKeyAlgo, query.CreationTime},
		},
		Input: in,
	}

	return match, q
}

func fixedEmailMatch(db *database.DB) (string, *query.ClientFSS) {
	match := "epflepflepflepflepflepflepflepfl"

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].UserId.Email = match
	}

	h := blake2b.Sum256([]byte(match))
	in := utils.ByteToBits(h[:16])
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId},
		Input: in,
	}

	return match, q
}

func fixedStartsWithMatch(db *database.DB) (string, *query.ClientFSS) {
	match := "START"

	for i := 0; i < 50; i++ {
		newEmail := match + randomDB.KeysInfo[i].UserId.Email[5:]
		randomDB.KeysInfo[i].UserId.Email = newEmail
	}

	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromStart: len(match)},
		Input: in,
	}

	return match, q
}

func fixedEndsWithMatch(db *database.DB) (string, *query.ClientFSS) {
	match := "END"

	for i := 0; i < 50; i++ {
		newEmail := randomDB.KeysInfo[i].UserId.Email[:len(randomDB.KeysInfo[i].UserId.Email)-len(match)] + match
		randomDB.KeysInfo[i].UserId.Email = newEmail
	}

	in := utils.ByteToBits([]byte(match))
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.UserId, FromEnd: len(match)},
		Input: in,
	}

	return match, q
}

func fixedPkaMatch(db *database.DB) (packet.PublicKeyAlgorithm, *query.ClientFSS) {
	match := packet.PubKeyAlgoRSA

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].PubKeyAlgo = match
	}

	in := utils.ByteToBits([]byte{byte(match)})
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.PubKeyAlgo},
		Input: in,
	}

	return match, q
}

func fixedCreationTimeMatch(db *database.DB) (time.Time, *query.ClientFSS) {
	match := time.Date(2009, time.November, 0, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].CreationTime = match
	}

	binaryMatch, err := match.MarshalBinary()
	if err != nil {
		panic(err)
	}
	in := utils.ByteToBits(binaryMatch)
	q := &query.ClientFSS{
		Info:  &query.Info{Target: query.CreationTime},
		Input: in,
	}

	return match, q
}
