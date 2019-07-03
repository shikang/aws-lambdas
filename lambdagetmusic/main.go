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
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("ap-southeast-1"))

type ErrorJson struct {
	ErrorMsg string `json:"error"`
}

type Music struct {
	Artist    string `json:"artist"`
	SongTitle string `json:"songTitle"`
}

func getArtistMusic(artist string) ([]Music, error) {
	filt := expression.Name("Artist").Equal(expression.Value(artist))
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
		TableName:                 aws.String("Music"),
	}

	// Make the DynamoDB Query API call
	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	musics := []Music{}
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &musics)
	if err != nil {
		return nil, err
	}

	return musics, nil
}

func getArtistMusicResponse(artist string) (events.APIGatewayProxyResponse, error) {
	musics, err := getArtistMusic(artist)
	if err != nil {
		apiResponse := generateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	responseBody, err := json.Marshal(musics)
	if err != nil {
		apiResponse := generateErrorResponse(err.Error(), http.StatusInternalServerError)
		return apiResponse, err
	}

	apiResponse := events.APIGatewayProxyResponse{Body: string(responseBody), StatusCode: http.StatusOK}
	return apiResponse, nil
}

func generateErrorResponse(err string, statusCode int) events.APIGatewayProxyResponse {
	errJSON := &ErrorJson{ErrorMsg: err}
	errBody, _ := json.Marshal(errJSON)
	apiResponse := events.APIGatewayProxyResponse{Body: string(errBody), StatusCode: statusCode}
	return apiResponse
}

func HandleGetMusicRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod == "POST" {
		queryJson := Music{}
		err := json.Unmarshal([]byte(request.Body), &queryJson)
		if err != nil {
			apiResponse := generateErrorResponse(err.Error(), http.StatusInternalServerError)
			return apiResponse, err
		}

		fmt.Print("[POST] Get music from artist: " + queryJson.Artist)
		return getArtistMusicResponse(queryJson.Artist)
	} else if request.HTTPMethod == "GET" {
		if artist, ok := request.QueryStringParameters["artist"]; ok {
			fmt.Print("[GET] Get music from artist: " + artist)
			return getArtistMusicResponse(artist)
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
	lambda.Start(HandleGetMusicRequest)
}
