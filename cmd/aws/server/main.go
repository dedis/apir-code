package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var database *database.DB

func main() {
	// prepare the database
	lambda.Start(HandleQuery)
}

// HandleRequest handles the lambda request.
func HandleQuery(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// b64 encoded query
	queryEncoded := req.Body

	// decode b64 query
	query, err := base64.StdEncoding.DecodeString(queryEncoded)
	if err != nil {
		log.Fatalf("impossible to decode the query: %v", err)
	}

	// run answer

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf(`{"answer": "%s"}`, body),
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
