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
)

const (
	oneB            = 8
	oneKB           = 1024 * oneB
	oneMB           = 1024 * oneKB
	testBlockLength = 64
	numIdentifiers  = 100000
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

func TestAvgQuery(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedAvgQueryMatch(db)

	retrieveComplex(t, randomDB, q, match, "TestAvgQuery")
}

func TestAvgQueryPIR(t *testing.T) {
	if randomDB == nil {
		initRandomDB()
	}

	match, q := fixedAvgQueryMatch(db)

	retrieveComplexPIR(t, randomDB, q, match, "TestAvgQueryPIR")
}

func fixedAvgQueryMatch(db *database.DB) (time.Time, *query.ClientFSS) {
	matchYear := time.Date(2019, 0, 0, 0, 0, 0, 0, time.UTC)

	info := &query.Info{
		And: true,
		Avg: true,
	}
	q := info.ToAvgClientFSS("")

	return matchYear, q
}

func fixedAndQueryMatch(db *database.DB) (interface{}, *query.ClientFSS) {
	matchYear := time.Date(2019, 0, 0, 0, 0, 0, 0, time.UTC)
	matchOrganization := ".edu"

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].CreationTime = matchYear
		originalEmail := randomDB.KeysInfo[i].UserId.Email
		lenOriginalEmail := len(originalEmail)
		newEmail := originalEmail[:lenOriginalEmail-len(matchOrganization)] + matchOrganization
		randomDB.KeysInfo[i].UserId.Email = newEmail
	}

	info := &query.Info{
		And:       true,
		FromStart: 0,
		FromEnd:   len(matchOrganization),
	}
	q := info.ToAndClientFSS(matchOrganization)

	return []interface{}{matchYear, matchOrganization}, q
}

func fixedEmailMatch(db *database.DB) (string, *query.ClientFSS) {
	match := "epflepflepflepflepflepflepflepfl"

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].UserId.Email = match
	}

	info := &query.Info{Target: query.UserId}
	q := info.ToEmailClientFSS(match)

	return match, q
}

func fixedStartsWithMatch(db *database.DB) (string, *query.ClientFSS) {
	match := "START"

	for i := 0; i < 50; i++ {
		newEmail := match + randomDB.KeysInfo[i].UserId.Email[5:]
		randomDB.KeysInfo[i].UserId.Email = newEmail
	}

	info := &query.Info{Target: query.UserId, FromStart: len(match)}
	q := info.ToEmailClientFSS(match)

	return match, q
}

func fixedEndsWithMatch(db *database.DB) (string, *query.ClientFSS) {
	match := "END"

	for i := 0; i < 50; i++ {
		newEmail := randomDB.KeysInfo[i].UserId.Email[:len(randomDB.KeysInfo[i].UserId.Email)-len(match)] + match
		randomDB.KeysInfo[i].UserId.Email = newEmail
	}

	info := &query.Info{Target: query.UserId, FromEnd: len(match)}
	q := info.ToEmailClientFSS(match)

	return match, q
}

func fixedPkaMatch(db *database.DB) (packet.PublicKeyAlgorithm, *query.ClientFSS) {
	match := packet.PubKeyAlgoRSA
	matchString := "RSA"

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].PubKeyAlgo = match
	}

	info := &query.Info{Target: query.PubKeyAlgo}
	q := info.ToPKAClientFSS(matchString)

	return match, q
}

func fixedCreationTimeMatch(db *database.DB) (time.Time, *query.ClientFSS) {
	match := time.Date(2009, 0, 0, 0, 0, 0, 0, time.UTC)
	matchString := "2009"

	for i := 0; i < 50; i++ {
		randomDB.KeysInfo[i].CreationTime = match
	}

	info := &query.Info{Target: query.CreationTime}
	q := info.ToCreationTimeClientFSS(matchString)

	return match, q
}
