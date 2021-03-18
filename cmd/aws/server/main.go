package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(HandleRequest)
}

// HandleRequest handles the lambda request.
func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// b64 encoded query
	query := req.Body

	// decode b64 query

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf(`{"query": "%s"}`, body),
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
