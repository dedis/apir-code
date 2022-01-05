package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/si-co/vpir-code/cmd/grpc/client/manager"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

var grpcOpts = []grpc.CallOption{
	grpc.UseCompressor(gzip.Name),
	grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
	grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
}

// main starts an interactive CLI to perform queries.
func main() {
	configPath := os.Getenv("VPIR_CONFIG")
	if configPath == "" {
		log.Fatalf("Please provide VPIR_CONFIG as env variable")
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	manager := manager.NewManager(*config, grpcOpts)

	actor, err := manager.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	// the initial questions: get a key or some stats ?
	prompt := &survey.Select{
		Message: "What do you want to do ?",
		Options: []string{"📦 Download a key", "🔎 Get stats", "👉 exit"},
	}

	var action string

	for {
		// perform the questions
		err = survey.AskOne(prompt, &action)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		switch action {
		case "📦 Download a key":
			err = downloadKey(actor)
			if err != nil {
				log.Fatalf("failed to download key: %v", err)
			}

		case "🔎 Get stats":
			err = getStats(actor)
			if err != nil {
				log.Fatalf("failed to get stats: %v", err)
			}

		case "👉 exit":
			fmt.Println("bye 👋")
			os.Exit(0)
		}
	}
}

// downloadKey performs a simple query on an email
func downloadKey(actor manager.Actor) error {
	prompt := &survey.Input{Message: "Enter the email"}

	var email string

	err := survey.AskOne(prompt, &email)
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	fmt.Println("email is", email)

	dbInfo, err := actor.GetDBInfos()
	if err != nil {
		return xerrors.Errorf("failed to get db info: %v", err)
	}

	client := client.NewPIR(utils.RandomPRG(), &dbInfo[0])

	result, err := actor.GetKey(email, dbInfo[0], client)
	if err != nil {
		return xerrors.Errorf("failed to get result: %v", err)
	}

	fmt.Println("Result:", result)

	return nil
}

// getStats either counts emails or computes an average lifetime
func getStats(actor manager.Actor) error {
	prompt := &survey.Select{
		Message: "What kind of stat do you want ?",
		Options: []string{"✌️ Count emails", "📆 Get average lifetime based on email"},
	}

	var answer string

	err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required))
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	switch answer {
	case "✌️ Count emails":
		err = countStats(actor)
		if err != nil {
			return xerrors.Errorf("failed to perform count stats: %v", err)
		}
	case "📆 Get average lifetime based on email":
		err = getAvg(actor)
		if err != nil {
			return xerrors.Errorf("failed to get avg: %v", err)
		}
	}

	return nil
}

// countStats counts emails based on different attributes
func countStats(actor manager.Actor) error {
	prompt := &survey.Select{
		Message: "What attribute do you want to count on ?",
		Options: []string{"email", "algo", "creation"},
	}

	var answer string

	err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required))
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	var clientQuery *query.ClientFSS
	var queryString string

	switch answer {
	case "email":
		clientQuery, queryString, err = getCountByEmailQuery()
	case "algo":
		clientQuery, queryString, err = getCountByAlgoQuery()
	case "creation":
		clientQuery, queryString, err = getCountByCreationQuery()
	}

	if err != nil {
		return xerrors.Errorf("failed to get stats: %v", err)
	}

	count, err := executeStatsQuery(clientQuery, actor)
	if err != nil {
		return xerrors.Errorf("failed to execute count query: %v", err)
	}

	var res bool
	prompt2 := &survey.Input{
		Message: fmt.Sprintf("Result of %s: %d\n. Type enter to continue.", queryString, count),
	}

	survey.AskOne(prompt2, &res)

	return nil
}

func getCountByEmailQuery() (*query.ClientFSS, string, error) {
	var prompt survey.Prompt

	prompt = &survey.Select{
		Message: "Select the kind of query",
		Options: []string{"begin with", "end with"},
	}

	var position string

	err := survey.AskOne(prompt, &position)
	if err != nil {
		return nil, "", xerrors.Errorf("failed to ask: %v", err)
	}

	prompt = &survey.Input{Message: "Enter your query text"}

	var q string

	err = survey.AskOne(prompt, &q)
	if err != nil {
		return nil, "", xerrors.Errorf("failed to ask: %v", err)
	}

	fromStart, fromEnd := len(q), 0

	if position == "end with" {
		fromStart, fromEnd = 0, len(q)
	}

	info := &query.Info{
		Target:    query.UserId,
		FromStart: fromStart,
		FromEnd:   fromEnd,
		And:       false,
	}

	clientQuery := info.ToEmailClientFSS(q)
	queryString := fmt.Sprintf("all emails that %s '%s'", position, q)

	return clientQuery, queryString, nil
}

func getCountByAlgoQuery() (*query.ClientFSS, string, error) {
	var prompt survey.Prompt

	prompt = &survey.Select{
		Message: "Select the kind of algo",
		Options: []string{"RSA", "ElGamal", "DSA", "ECDH", "ECDSA"},
	}

	var algo string

	err := survey.AskOne(prompt, &algo)
	if err != nil {
		return nil, "", xerrors.Errorf("failed to ask: %v", err)
	}

	info := &query.Info{
		Target: query.PubKeyAlgo,
	}

	clientQuery := info.ToPKAClientFSS(algo)
	queryString := fmt.Sprintf("all emails that uses the '%s' algorithm", algo)

	return clientQuery, queryString, nil
}

func getCountByCreationQuery() (*query.ClientFSS, string, error) {
	var prompt survey.Prompt

	prompt = &survey.Input{Message: "Enter a year"}

	var yearStr string

	intValidator := func(ans interface{}) error {
		str := ans.(string)

		year, err := strconv.Atoi(str)
		if err != nil || year < 0 {
			return xerrors.Errorf("please enter a positive integer")
		}

		return nil
	}

	err := survey.AskOne(prompt, &yearStr, survey.WithValidator(intValidator))
	if err != nil {
		return nil, "", xerrors.Errorf("failed to ask: %v", err)
	}

	info := &query.Info{
		Target: query.CreationTime,
	}

	clientQuery := info.ToCreationTimeClientFSS(yearStr)
	queryString := fmt.Sprintf("all emails created before year '%s'", yearStr)

	return clientQuery, queryString, nil
}

func getAvg(actor manager.Actor) error {
	var prompt survey.Prompt

	prompt = &survey.Select{
		Message: "Select the kind of query",
		Options: []string{"begin with", "end with"},
	}

	var position string

	err := survey.AskOne(prompt, &position)
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	prompt = &survey.Input{Message: "Enter your query text"}

	var q string

	err = survey.AskOne(prompt, &q)
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	fromStart, fromEnd := len(q), 0

	if position == "end with" {
		fromStart, fromEnd = 0, len(q)
	}

	info := &query.Info{
		FromStart: fromStart,
		FromEnd:   fromEnd,
		And:       true,
		Avg:       true,
	}

	clientQuery := info.ToAvgClientFSS(q)
	queryString := fmt.Sprintf("average lifetime of all emails that %s '%s'", position, q)

	count, err := executeStatsQuery(clientQuery, actor)
	if err != nil {
		return xerrors.Errorf("failed to execute count query: %v", err)
	}

	var res bool
	prompt2 := &survey.Input{
		Message: fmt.Sprintf("Result of %s: %d\n. Type enter to continue.", queryString, count),
	}

	survey.AskOne(prompt2, &res)

	return nil
}

// executeStatsQuery takes a client query and executes it
func executeStatsQuery(clientQuery *query.ClientFSS, actor manager.Actor) (uint32, error) {
	in, err := clientQuery.Encode()
	if err != nil {
		return 0, xerrors.Errorf("failed to encode query: %v", err)
	}

	dbInfo, err := actor.GetDBInfos()
	if err != nil {
		return 0, xerrors.Errorf("failed to get db info: %v", err)
	}

	client := client.NewPredicateAPIR(utils.RandomPRG(), &dbInfo[0])

	queries, err := client.QueryBytes(in, len(dbInfo))
	if err != nil {
		return 0, xerrors.Errorf("failed to query bytes: %v", err)
	}

	answers := actor.RunQueries(queries)

	result, err := client.ReconstructBytes(answers)
	if err != nil {
		return 0, xerrors.Errorf("failed to reconstruct bytes: %v", err)
	}

	count, ok := result.(uint32)
	if !ok {
		return 0, xerrors.Errorf("failed to cast result, wrong type %T", result)
	}

	return count, nil
}
