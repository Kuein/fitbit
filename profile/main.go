package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"net/url"
	"text/template"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var f embed.FS

//go:embed profile.html
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
	Defence    int    `dynamodbav: "defence"`
	HP         int    `dynamodbav: "hp"`
	Username   string `dynamodbav: "primary"`
	Picture    string `dynamodbav: "picture"`
	FitBit     bool
}

var svc = dynamodb.New(session.New())

type Session struct {
	SessionId string `dynamodbav:"primary"`
	Username  string `dynamodbav:"username"`
}

type Character struct {
	Experience string `json:"exp"`
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

func handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	fmt.Println(request)
	username := getUserNameFromSession(request.Cookies[0])
	if username == "" {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 401,
			Headers:    map[string]string{"Location": "/Prod/login"},
		}, nil
	}
	switch request.RouteKey {
	case "GET /profile":
		{

			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(username)},
					"secondary": {S: aws.String("profile")},
				},
				TableName: aws.String("Fitbit"),
			}
			result, err := svc.GetItem(input)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 301,
					Headers:    map[string]string{"Location": ErrorPage},
				}, err
			}
			actual := Profile{}
			err = dynamodbattribute.UnmarshalMap(result.Item, &actual)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 500,
				}, err
			}
			var buf bytes.Buffer
			t, _ := template.New("profile").Parse(profilePage)
			_, actual.FitBit = result.Item["fitbit_auth"]
			t.Execute(&buf, actual)
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
				Body: buf.String(),
			}, nil

		}
	default:
		{
			arr, _ := base64.StdEncoding.DecodeString(request.Body)
			fmt.Println(arr)
			vals, _ := url.ParseQuery(string(arr))
			fmt.Println(vals)
			update := map[string]string{
				"exp": vals["experience"][0],
				"att": vals["attack"][0],
				"def": vals["defence"][0],
				"hp":  vals["hp"][0],
			}
			fmt.Println(update)
			update_input := &dynamodb.UpdateItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(username)},
					"secondary": {S: aws.String("profile")},
				},
				UpdateExpression: aws.String("set experience = :experience, attack = :att, defence = :def, hp = :hp"),
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":experience": {N: aws.String(update["exp"])},
					":att":        {N: aws.String(update["att"])},
					":def":        {N: aws.String(update["def"])},
					":hp":         {N: aws.String(update["hp"])},
				},
				TableName: aws.String("Fitbit"),
			}
			_, err := svc.UpdateItem(update_input)
			if err != nil {
				fmt.Printf("Update profile XP failed: %v\n", err)
			}

			return events.APIGatewayV2HTTPResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": "/profile"},
			}, nil

		}

	}
}

func main() {
	lambda.Start(handler)
}
