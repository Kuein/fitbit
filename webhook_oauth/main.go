package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	clientId     = "23BMSG"
	clientSecret = "dabf4bc970ad5d9d709f096da275000b"
	redirect     = "https://1wrme4wa94.execute-api.eu-central-1.amazonaws.com/webhook"
)

var svc = dynamodb.New(session.New())

type Auth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"user_id"`
	Scope        string `json:"scope"`
}
type Session struct {
	SessionId string `dynamodbav:"primary"`
	Username  string `dynamodbav:"username"`
}

func getUserNameFromSession(sess string) string {
	fmt.Println(sess)

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
	switch request.RawPath {
	case "/auth_code":
		{
			// return redirect to fitbit auth page
			url, err := http.NewRequest("GET", "https://www.fitbit.com/oauth2/authorize", nil)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{
					Body:       "Cannot create requests",
					StatusCode: 500,
				}, nil
			}
			q := url.URL.Query()
			q.Add("client_id", clientId)
			q.Add("scope", "settings weight profile nutrition activity sleep")
			q.Add("redirect_url", redirect)
			q.Add("response_type", "code")
			url.URL.RawQuery = q.Encode()
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": url.URL.String()},
			}, nil

		}
	default:
		username := getUserNameFromSession(request.Cookies[0])
		if username == "" {

			return events.APIGatewayV2HTTPResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": "/Prod/login"},
			}, nil
		}
		{
			// parse GET url and get CODE from URL
			code := request.QueryStringParameters["code"]
			if len(code) == 0 {
				return events.APIGatewayV2HTTPResponse{
					Body:       "No code specified",
					StatusCode: 400,
				}, nil
			}

			// send post request to FitBit for generation of auth data
			basic := base64.URLEncoding.EncodeToString([]byte(clientId + ":" + clientSecret))

			v := url.Values{}
			v.Set("client_id", clientId)
			v.Set("grant_type", "authorization_code")
			v.Set("code", code)
			v.Set("redirect_url", redirect)
			url, err := http.NewRequest("POST", "https://api.fitbit.com/oauth2/token", strings.NewReader(v.Encode()))
			if err != nil {
				return events.APIGatewayV2HTTPResponse{}, err
			}
			url.Header.Add("Authorization", "Basic "+basic)
			url.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			c := &http.Client{}
			res, err := c.Do(url)
			if err != nil {
				fmt.Printf("http.Do() error: %v\n", err)
				return events.APIGatewayV2HTTPResponse{}, err
			}
			defer res.Body.Close()

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("ioutil.ReadAll() error: %v\n", err)
				return events.APIGatewayV2HTTPResponse{}, err
			}

			auth := Auth{}
			err = json.Unmarshal(data, &auth)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{}, err
			}
			av, err := dynamodbattribute.MarshalMap(auth)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{}, err
			}
			// Username???

			input := &dynamodb.UpdateItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"primary":   {S: aws.String(username)},
					"secondary": {S: aws.String("profile")},
				},
				UpdateExpression: aws.String("set fitbit_auth = :fitbit_auth, user_id = :user_id"),
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":fitbit_auth": {M: av},
					":user_id":     {S: aws.String(auth.UserId)},
				},
				TableName: aws.String("Fitbit"),
			}
			_, err = svc.UpdateItem(input)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{}, err
			}
			// create subscription
			url, err = http.NewRequest("POST", "https://api.fitbit.com/1/user/"+auth.UserId+"/activities/apiSubscriptions/"+auth.UserId+".json", nil)
			url.Header.Add("Authorization", "Bearer "+auth.AccessToken)
			url.Header.Add("Content-Type", "application/json")
			res, err = c.Do(url)
			if err != nil {
				fmt.Println(err)
				return events.APIGatewayV2HTTPResponse{}, err
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
