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
	uuid "github.com/satori/go.uuid"
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

// To match column names
type Item struct {
	ID        string `json:"ID"`
	Title     string `json:"Title"`
	Completed bool   `json:"Completed"`
}

func GenerateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "OPTIONS,POST",
		},
		Body:       string(errBody),
		StatusCode: statusCode}
	return apiResponse
}

func AddTodo(todo Todos) (events.APIGatewayProxyResponse, error) {
	id, err := uuid.NewV4()
	if err != nil {
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	idStr := id.String()
	fmt.Println("New uuid: " + idStr)
	todo.ID = idStr

	item := Item{
		ID:        todo.ID,
		Title:     todo.Title,
		Completed: todo.Completed,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		fmt.Println("Got error marshalling new todo item:")
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	todoByte, err := json.Marshal(item)
	if err == nil {
		fmt.Println(string(todoByte))
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("Todos"),
	}

	_, err = db.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem")
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	responseBody, err := json.Marshal(todo)
	if err != nil {
		apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	apiResponse := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "OPTIONS,POST",
		},
		Body:       string(responseBody),
		StatusCode: http.StatusOK}
	return apiResponse, nil
}

func HandleAddTodosRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "POST" {
		newTodo := Todos{}
		err := json.Unmarshal([]byte(request.Body), &newTodo)
		if err != nil {
			apiResponse := GenerateErrorResponse(err.Error(), http.StatusInternalServerError)
			return apiResponse, err
		}

		return AddTodo(newTodo)
	} else {
		err := errors.New("Method not allowed")
		apiResponse := GenerateErrorResponse("Method Not OK", http.StatusBadGateway)
		return apiResponse, err
	}
}

func main() {
	lambda.Start(HandleAddTodosRequest)
}
