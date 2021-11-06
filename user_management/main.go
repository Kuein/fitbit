package main

import (
	"embed"
	"fmt"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var f embed.FS

// TODO: ProfilePage and ErronPage
//go:embed signup.page.tmpl
var signupPage string

//go:embed login.page.tmpl
var loginPage string

const (
	tableName = "Fitbit"
	ErrorPage = ""
)

type User struct {
	Username string `json:"username" dynamodbav:"primary"`
	Password string `json:"password" dynamodbav:"password"`
}

var svc = dynamodb.New(session.New())

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	pageKey := request.HTTPMethod + " " + request.Path
	switch pageKey {
	case "GET /signup":
		{
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
				Body: signupPage,
			}, nil
		}
	case "GET /login":
		{
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
				Body: loginPage,
			}, nil
		}
	case "POST /signup":
		// TODO: hash password
		// TODO: check users for existence
		// TODO: confirm user somehow
		{
			vals, _ := url.ParseQuery(request.Body)
			userData := User{
				Username: vals["username"][0],
				Password: vals["password"][0],
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
				Headers:    map[string]string{"Location": "/Prod/profile/" + userData.Username},
			}, nil

		}
	case "POST /login":
		// TODO: hash password
		// TODO: implement 2FA
		{
			vals, _ := url.ParseQuery(request.Body)
			userData := User{
				Username: vals["username"][0],
				Password: vals["password"][0],
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
				Headers:    map[string]string{"Location": "/Prod/profile/" + userData.Username},
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
