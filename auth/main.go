package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var svc = dynamodb.New(session.New())

type Session struct {
	SessionId string `dynamodbav:"primary"`
	Username  string `dynamodbav:"username"`
}

func getUserNameFromSession(sess string) string {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"primary":   {S: aws.String(sess)},
			"secondary": {S: aws.String("session")},
		},
		TableName: aws.String("Fitbit"),
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return ""
	}
	actual := Session{}
	dynamodbattribute.UnmarshalMap(result.Item, &actual)

	return actual.Username
}

func handler(request events.APIGatewayV2HTTPRequest) (map[string]bool, error) {
	fmt.Println(request)
	username := getUserNameFromSession(request.Headers["Cookie"])
	if username == "" {
		return map[string]bool{"isAuthorized": false}, nil
	}
	return map[string]bool{"isAuthorized": true}, nil
}

func main() {
	lambda.Start(handler)
}
