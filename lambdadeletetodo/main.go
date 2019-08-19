package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-southeast-1"))

type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

type Todos struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func GenerateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "OPTIONS,POST,DELETE",
		},
		Body:       string(errBody),
		StatusCode: statusCode}
	return apiResponse
}

func DeleteTodo(todo Todos) (events.APIGatewayProxyResponse, error) {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(todo.ID),
			},
			"Title": {
				S: aws.String(todo.Title),
			},
		},
		TableName: aws.String("Todos"),
	}

	_, err := db.DeleteItem(input)
	if err != nil {
		fmt.Println("Got error calling DeleteItem")
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "OPTIONS,POST,DELETE",
		},
		Body:       "{ \"success\": true }",
		StatusCode: http.StatusOK}
	return apiResponse, nil
}

func GetTodosByID(val string, limit int64) ([]Todos, error) {
	// Build the query input parameters
	params := &dynamodb.QueryInput{
		TableName: aws.String("Todos"),
		KeyConditions: map[string]*dynamodb.Condition{
			"ID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(val),
					},
				},
			},
		},
	}

	// Make the DynamoDB Query API call
	result, err := db.Query(params)
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

func HandleDeleteTodoRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "POST" || request.HTTPMethod == "DELETE" {
		delTodo := Todos{}

		err := json.Unmarshal([]byte(request.Body), &delTodo)
		if err != nil {
			apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
			return apiResponse, err
		}

		if delTodo.ID != "" && delTodo.ID != "null" {
			if delTodo.Title == "" || delTodo.Title == "null" {
				todos, err := GetTodosByID(delTodo.ID, 1)

				if err != nil {
					apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
					return apiResponse, err
				}
				delTodo.Title = todos[0].Title
			}

			fmt.Println("Deleting: " + delTodo.ID + " - " + delTodo.Title)
			return DeleteTodo(delTodo)
		} else {
			err := errors.New("Deleting ID not specified")
			apiResponse := GenerateErrorResponse(err.Error(), http.StatusBadGateway)
			return apiResponse, err
		}
	} else {
		err := errors.New("Method not allowed")
		apiResponse := GenerateErrorResponse("Method Not OK", http.StatusBadGateway)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleDeleteTodoRequest)
}
