package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/si-co/vpir-code/cmd/grpc/client/manager"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// main start an interactive CLI to perform queries.
func main() {
	configPath := os.Getenv("VPIR_CONFIG")
	if configPath == "" {
		log.Fatalf("Please provide VPIR_CONFIG as env variable")
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// the initial questions: complex or simple query ?
	prompt := &survey.Select{
		Message: "What do you want to do ?",
		Options: []string{"Download a key", "Get stats", "exit"},
		Default: "Download a key",
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
		case "Download a key":
			err = downloadKey(*config)
			if err != nil {
				log.Fatalf("failed to download key: %v", err)
			}

		case "Get stats":
			err = getStats(*config)
			if err != nil {
				log.Fatalf("failed to get stats: %v", err)
			}

		case "exit":
			fmt.Println("bye 👋")
			os.Exit(0)
		}
	}
}

// downloadKey performs a simple query on an email
func downloadKey(config utils.Config) error {
	prompt := &survey.Input{Message: "Enter the email"}

	var email string

	err := survey.AskOne(prompt, &email)
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	fmt.Println("email is", email)

	opts := []grpc.CallOption{
		grpc.UseCompressor(gzip.Name),
		grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
		grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
	}

	manager := manager.NewManager(config, opts)

	err = manager.Connect()
	if err != nil {
		return xerrors.Errorf("failed to connect: %v", err)
	}

	dbInfo, err := manager.GetDBInfos()
	if err != nil {
		return xerrors.Errorf("failed to get db info: %v", err)
	}

	client := client.NewPIR(utils.RandomPRG(), &dbInfo[0])

	result, err := manager.GetKey(email, dbInfo[0], client)
	if err != nil {
		return xerrors.Errorf("failed to get result: %v", err)
	}

	fmt.Println("Result:", result)

	return nil
}

func getStats(config utils.Config) error {
	prompt := &survey.Select{
		Message: "Kind of stat do you want ?",
		Options: []string{"Count emails", "Get everage lifetime"},
		Default: "Count emails",
	}

	var answer string

	err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required))
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	switch answer {
	case "Count emails":
		err = countStats(config)
		if err != nil {
			log.Fatalf("failed to get stats email: %v", err)
		}
	case "Get everage lifetime":
	}

	return nil
}

func countStats(config utils.Config) error {
	prompt := &survey.Select{
		Message: "What attribute do you want to count on ?",
		Options: []string{"email", "algo", "creation"},
		Default: "email",
	}

	var answer string

	err := survey.AskOne(prompt, &answer, survey.WithValidator(survey.Required))
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	switch answer {
	case "email":
		err = getStatsEmail(config)
		if err != nil {
			log.Fatalf("failed to get stats email: %v", err)
		}
	case "algo":
	case "creation":
	}

	return nil
}

func getStatsEmail(config utils.Config) error {
	var prompt survey.Prompt

	prompt = &survey.Select{
		Message: "Select the kind of query",
		Options: []string{"begin with", "end with"},
		Default: "begins with",
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
		Target:    query.UserId,
		FromStart: fromStart,
		FromEnd:   fromEnd,
		And:       false,
	}

	clientQuery := info.ToEmailClientFSS(q)

	in, err := clientQuery.Encode()
	if err != nil {
		return xerrors.Errorf("failed to encode query: %v", err)
	}

	opts := []grpc.CallOption{
		grpc.UseCompressor(gzip.Name),
		grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
		grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
	}

	manager := manager.NewManager(config, opts)

	err = manager.Connect()
	if err != nil {
		return xerrors.Errorf("failed to connect: %v", err)
	}

	dbInfo, err := manager.GetDBInfos()
	if err != nil {
		return xerrors.Errorf("failed to get db info: %v", err)
	}

	client := client.NewPredicateAPIR(utils.RandomPRG(), &dbInfo[0])

	queries, err := client.QueryBytes(in, len(dbInfo))
	if err != nil {
		return xerrors.Errorf("failed to query bytes: %v", err)
	}

	answers := manager.RunQueries(queries)

	result, err := client.ReconstructBytes(answers)
	if err != nil {
		return xerrors.Errorf("failed to reconstruct bytes: %v", err)
	}

	count, ok := result.(uint32)
	if !ok {
		return xerrors.Errorf("failed to cast result, wrong type %T", count)
	}

	var res bool
	prompt = &survey.Input{
		Message: fmt.Sprintf("Result of all emails that %s '%s': %d\n. Type enter to continue.", position, q, count),
	}

	survey.AskOne(prompt, &res)

	return nil
}
