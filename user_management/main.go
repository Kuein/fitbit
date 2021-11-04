package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// TODO: ProfilePage and ErronPage

const (
	tableName   = "Fitbit"
	ProfilePage = ""
	ErrorPage   = ""
)

type User struct {
	Username string `json:"username" dynamodbav:"primary"`
	Password string `json:"password" dynamodbav:"password"`
}

var svc = dynamodb.New(session.New())

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: render signup page
	// TODO: render login page
	switch request.Path {
	case "/signup":
		// TODO: hash password
		// TODO: check users for existence
		// TODO: confirm user somehow
		{
			userData := User{}
			if err := json.Unmarshal([]byte(request.Body), &userData); err != nil {
				return events.APIGatewayProxyResponse{}, err
			}
			input := &dynamodb.PutItemInput{
				Item: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(userData.Username)},
					"secondary": {S: aws.String("profile")},
					"password":  {S: aws.String(userData.Password)},
				},
				TableName: aws.String("Fitbit"),
			}
			_, err := svc.PutItem(input)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: 301,
					Headers:    map[string]string{"Location": ErrorPage},
				}, err
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": ProfilePage},
			}, nil

		}
	case "/login":
		// TODO: hash password
		// TODO: implement 2FA
		{
			userData := User{}
			if err := json.Unmarshal([]byte(request.Body), &userData); err != nil {
				return events.APIGatewayProxyResponse{}, err
			}
			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(userData.Username)},
					"secondary": {S: aws.String("profile")},
				},
				TableName: aws.String("Fitbit"),
			}
			result, err := svc.GetItem(input)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: 301,
					Headers:    map[string]string{"Location": ErrorPage},
				}, err
			}
			actual := User{}
			err = dynamodbattribute.UnmarshalMap(result.Item, &actual)

			if actual.Password != userData.Password {
				return events.APIGatewayProxyResponse{
					StatusCode: 400,
					Body:       "Wrong password",
				}, err
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": ProfilePage},
			}, nil

		}
	default:
		{

			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("What exactly you want?"),
				StatusCode: 200,
			}, nil

		}
	}
}

func main() {
	lambda.Start(handler)
}
