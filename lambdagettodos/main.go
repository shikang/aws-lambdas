package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-southeast-1"))

type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

type Todos struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func getTodosWithoutAnyFilters(limit int64) ([]Todos, error) {
	// Build the scan input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String("Todos"),
		Limit:     aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	todos := []Todos{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &todos)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

func getTodosByCompleted(val bool, limit int64) ([]Todos, error) {
	filt := expression.Name("Completed").Equal(expression.Value(val))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String("Todos"),
		Limit:                     aws.Int64(limit),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	todos := []Todos{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &todos)
	if err != nil {
		return nil, err
	}

	return todos, nil
}

func getTodos(filter string, val string, limit int64) ([]Todos, error) {
	if val == "any" {
		return getTodosWithoutAnyFilters(limit)
	}

	switch filter {
	case "completed":
		completed, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return getTodosByCompleted(completed, limit)
	default:
		err := errors.New("Invalid filter")
		return nil, err
	}
}

func getTodosResponse(filters string, val string, limit int64) (events.APIGatewayProxyResponse, error) {
	todos, err := getTodos(filters, val, limit)
	if err != nil {
		apiResponse := generateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	responseBody, err := json.Marshal(todos)
	if err != nil {
		apiResponse := generateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body:       string(responseBody),
		StatusCode: http.StatusOK}
	return apiResponse, nil
}

func generateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body:       string(errBody),
		StatusCode: statusCode}
	return apiResponse
}

func HandleGetTodosRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "GET" {
		if completed, ok := request.QueryStringParameters["completed"]; ok {
			var queryLimit int64 = 10
			if limit, ok := request.QueryStringParameters["limit"]; ok {
				queryLimit, _ = strconv.ParseInt(limit, 10, 64)
			}

			fmt.Print("[GET] Get todos with completed filter: " + completed)
			return getTodosResponse("completed", completed, queryLimit)
		} else {
			err := errors.New("Empty query string")
			apiResponse := generateErrorResponse("Empty query string", http.StatusBadGateway)
			return apiResponse, err
		}
	} else {
		err := errors.New("Method not allowed")
		apiResponse := generateErrorResponse("Method Not OK", http.StatusBadGateway)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleGetTodosRequest)
}
