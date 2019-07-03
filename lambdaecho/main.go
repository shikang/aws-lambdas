package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

type EchoJson struct {
	Payload     string    `json:"payload"`
	Timestamp   time.Time `json:"timestamp"`
	RequestType string    `json:"request"`
}

func generateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{Body: string(errBody), StatusCode: statusCode}
	return apiResponse
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "POST" {
		echoJSON := EchoJson{}
		json.Unmarshal([]byte(request.Body), &echoJSON)
		echoJSON.Timestamp = time.Now().Local()
		echoJSON.RequestType = request.HTTPMethod

		reponseBody, err := json.Marshal(echoJSON)
		if err != nil {
			err := errors.New("Marshal Json Error")
			apiResponse := generateErrorResponse(err.Error(), 500)
			return apiResponse, err
		}
		apiResponse := events.APIGatewayProxyResponse{Body: string(reponseBody), StatusCode: 200}
		return apiResponse, nil
	} else {
		err := errors.New("Method not allowed")
		apiResponse := generateErrorResponse("Method Not OK", 502)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleRequest)
}
