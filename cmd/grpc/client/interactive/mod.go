package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/si-co/vpir-code/cmd/grpc/client/manager"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

func main() {
	configPath := os.Getenv("VPIR_CONFIG")
	if configPath == "" {
		log.Fatalf("Please provide VPIR_CONFIG as env variable")
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// the initial questions
	var qs = []*survey.Question{
		{
			Name: "action",
			Prompt: &survey.Select{
				Message: "What do you want to do ?",
				Options: []string{"Query a simple email address", "Make a complex query", "exit"},
				Default: "Query a simple email address",
			},
		},
	}

	// the answers will be written to this struct
	answers := struct {
		Scheme string // survey will match the question and field names
		Action string
	}{}

	for {
		// perform the questions
		err = survey.Ask(qs, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if answers.Action == "Query a simple email address" {
			simpleQuery(*config)
		}

		if answers.Action == "exit" {
			fmt.Println("bye 👋")
			os.Exit(0)
		}
	}
}

func simpleQuery(config utils.Config) error {
	// the simple question
	var qs = []*survey.Question{
		{
			Name:   "email",
			Prompt: &survey.Input{Message: "Which email do you want?"},
		},
	}

	answers := struct {
		Email string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		return xerrors.Errorf("failed to ask: %v", err)
	}

	fmt.Println("email is", answers.Email)

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

	result, err := manager.GetKey(answers.Email, dbInfo[0], client)
	if err != nil {
		return xerrors.Errorf("failed to get result: %v", err)
	}

	fmt.Println("Result:", result)

	return nil
}

func complexQuery() {

}
