package main

import (
	"embed"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var f embed.FS

// TODO: ProfilePage and ErronPage
//go:embed signup.html
var signupPage string

//go:embed login.html
var loginPage string
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

const (
	tableName = "Fitbit"
	ErrorPage = ""
)

type User struct {
	Username string `json:"username" dynamodbav:"primary"`
	Password string `json:"password" dynamodbav:"password"`
}

var svc = dynamodb.New(session.New())

func setSession(user User) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	token := string(b)

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"primary":   {S: aws.String(token)},
			"secondary": {S: aws.String("session")},
			"username":  {S: aws.String(user.Username)},
		},
		TableName: aws.String("Fitbit"),
	}

	svc.PutItem(input)
	return token
}

func decode64(str string) User {
	arr, _ := base64.StdEncoding.DecodeString(str)
	mm, _ := url.ParseQuery(string(arr))
	return User{
		Username: mm["username"][0],
		Password: mm["password"][0],
	}
}

func handler(request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	pageKey := request.RouteKey
	switch pageKey {
	case "GET /signup":
		{
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "text/html; charset=UTF-8",
				},
				Body: signupPage,
			}, nil
		}
	case "GET /login":
		{
			return events.APIGatewayV2HTTPResponse{
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
			userData := decode64(request.Body)
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
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 301,
					Headers:    map[string]string{"Location": ErrorPage},
				}, err
			}

			return events.APIGatewayV2HTTPResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": "/profile", "Set-Cookie": setSession(userData)},
			}, nil

		}
	case "POST /login":
		// TODO: hash password
		// TODO: implement 2FA
		{
			userData := decode64(request.Body)
			input := &dynamodb.GetItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(userData.Username)},
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
			actual := User{}
			err = dynamodbattribute.UnmarshalMap(result.Item, &actual)

			if actual.Password != userData.Password {
				return events.APIGatewayV2HTTPResponse{
					StatusCode: 400,
					Body:       "Wrong password",
				}, err
			}

			return events.APIGatewayV2HTTPResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": "/profile", "Set-Cookie": setSession(userData)},
			}, nil

		}
	default:
		{

			return events.APIGatewayV2HTTPResponse{
				Body:       fmt.Sprintf("What exactly you want?"),
				StatusCode: 200,
			}, nil

		}
	}
}

func main() {
	lambda.Start(handler)
}
