package main

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var f embed.FS

//go:embed profile.page.tmpl.html
var profilePage string

const (
	ErrorPage = ""
)

type AuthToken struct {
	Token   string
	Usename string
}

type Profile struct {
	Experience int    `dynamodbav: "experience"`
	Attack     int    `dynamodbav: "attack"`
	Defense    int    `dynamodbav: "defence"`
	HP         int    `dynamodbav: "hp"`
	Username   string `dynamodbav: "primary"`
	Picture    string `dynamodbav: "picture"`
	FitBit     bool
}

var svc = dynamodb.New(session.New())

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case "GET":
		{
			username := request.PathParameters["username"]

			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(username)},
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
			actual := Profile{}
			err = dynamodbattribute.UnmarshalMap(result.Item, &actual)
			if err != nil {
				return events.APIGatewayProxyResponse{
					StatusCode: 500,
				}, err
			}
			var buf bytes.Buffer
			t, _ := template.New("profile").Parse(profilePage)
			_, actual.FitBit = result.Item["fitbit_auth"]
			t.Execute(&buf, actual)
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
				Body: buf.String(),
			}, nil

		}
	default:
		{
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "Not implemented",
			}, nil

		}

	}
}

func main() {
	lambda.Start(handler)
}
