package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var svc = dynamodb.New(session.New())

const (
	client_id     = "23BMSG"
	client_secret = "dabf4bc970ad5d9d709f096da275000b"
	refresh_url   = "https://api.fitbit.com/oauth2/token"
)

var TableName = aws.String("Fitbit")

type Auth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"user_id"`
	Scope        string `json:"scope"`
}
type ActivityStat struct {
	Light  int `json:"lightlyActiveMinutes"`
	Medium int `json:"fairlyActiveMinutes"`
	Heavy  int `json:"veryActiveMinutes"`
}

type Daily struct {
	Summary ActivityStat `json:"summary"`
}

type Activity struct {
	Date   string `json:"date"`
	UserID string `json:"ownerId"`
}

type DayEXP struct {
	UserID string `dynamodbav: "primary"`
	Date   string `dynamodbav: "secondary"`
	EXP    int    `dynamodbav: "exp"`
}

type UserDetails struct {
	Username    string      `dynamodbav:"primary"`
	UserID      string      `dynamodbav:"user_id"`
	Credentials Credentials `dynamodbav:"fitbit_auth"`
}

type Credentials struct {
	AccessToken  string `dynamodbav:"access_token"`
	RefreshToken string `dynamodbav:"refresh_token"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case "GET":
		{
			vc := request.QueryStringParameters["verify"]
			if vc == "a2ffc806e7c6124256f430fdd384cae41c4a8482898c8460e66a07c3cfd27145" {
				return events.APIGatewayProxyResponse{
					StatusCode: 204,
				}, nil
			} else {
				return events.APIGatewayProxyResponse{
					StatusCode: 404,
				}, nil
			}
		}
	default:
		{
			basic := base64.URLEncoding.EncodeToString([]byte(client_id + ":" + client_secret))

			// find user by userID
			activities := make([]Activity, 0)
			err := json.Unmarshal([]byte(request.Body), &activities)
			if err != nil {
				fmt.Printf("Unable to marshall activity: %v\n", err)
				return events.APIGatewayProxyResponse{}, err
			}
			for _, act := range activities {
				scan_input := &dynamodb.ScanInput{
					FilterExpression: aws.String("user_id =:userid"),
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":userid": {S: aws.String(act.UserID)},
					},
					TableName: TableName,
				}
				result, err := svc.Scan(scan_input)
				if err != nil {
					fmt.Printf("Unable to scan : %v\n", err)
					return events.APIGatewayProxyResponse{}, err
				}
				if aws.Int64Value(result.Count) == 0 {
					fmt.Println("User " + act.UserID + " not found")
					continue
				} else if aws.Int64Value(result.Count) != 1 {
					fmt.Println("User ID " + act.UserID + " have dublicates")
					continue
				}
				c := &http.Client{}
				creds := UserDetails{}
				err = dynamodbattribute.UnmarshalMap(result.Items[0], &creds)
				if err != nil {
					fmt.Println("Credentials are wrong for user " + act.UserID)
					continue
				}
				// request activities from Activity.Date
				// refresh token
				v := url.Values{}
				v.Add("grant_type", "refresh_token")
				v.Add("refresh_token", creds.Credentials.RefreshToken)
				refresh, err := http.NewRequest("POST", refresh_url, strings.NewReader(v.Encode()))
				refresh.Header.Add("Authorization", "Basic "+basic)
				refresh.Header.Add("Content-Type", "application/x-www-form-urlencoded")

				res, err := c.Do(refresh)
				if err != nil {
					fmt.Printf("http.Do() error: %v\n", err)
					return events.APIGatewayProxyResponse{}, err
				}
				defer res.Body.Close()

				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					fmt.Printf("ioutil.ReadAll() error: %v\n", err)
					return events.APIGatewayProxyResponse{}, err
				}

				auth := Auth{}
				err = json.Unmarshal(data, &auth)
				if err != nil {
					return events.APIGatewayProxyResponse{}, err
				}
				av, err := dynamodbattribute.MarshalMap(auth)
				if err != nil {
					return events.APIGatewayProxyResponse{}, err
				}

				uinput := &dynamodb.UpdateItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"primary":   {S: aws.String(creds.Username)},
						"secondary": {S: aws.String("profile")},
					},
					UpdateExpression: aws.String("set fitbit_auth = :fitbit_auth"),
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":fitbit_auth": {M: av},
					},
					TableName: TableName,
				}
				_, err = svc.UpdateItem(uinput)
				aurl, _ := http.NewRequest("GET", "https://api.fitbit.com/1/user/-/activities/date/"+act.Date+".json", nil)
				aurl.Header.Add("Authorization", "Bearer "+auth.AccessToken)
				res, err = c.Do(aurl)
				if err != nil {
					fmt.Printf("http.Do() error: %v\n", err)
					return events.APIGatewayProxyResponse{}, err
				}
				defer res.Body.Close()
				dec := json.NewDecoder(res.Body)
				dailyData := Daily{}
				err = dec.Decode(&dailyData)
				if err != nil {
					fmt.Println(err)
					return events.APIGatewayProxyResponse{}, err
				}
				// calculate EXP for day
				dayEXP := dailyData.Summary.Light + dailyData.Summary.Medium*2 + dailyData.Summary.Heavy*5
				// compare EXP for day with already compared EXP from DynamoDB

				input := &dynamodb.GetItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"primary":   {S: aws.String(act.UserID)},
						"secondary": {S: aws.String(act.Date)},
					},
					TableName: TableName,
				}
				get_result, err := svc.GetItem(input)
				actual := DayEXP{}
				err = dynamodbattribute.UnmarshalMap(get_result.Item, &actual)
				if err != nil {
					fmt.Printf("Unmarshall result error: %v\n", err)
					continue
				}
				if actual.EXP == dayEXP {
					continue
				}
				delta := dayEXP - actual.EXP

				// change EXP in Profile for delta
				update_input := &dynamodb.UpdateItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"primary":   {S: aws.String(act.UserID)},
						"secondary": {S: aws.String(act.Date)},
					},
					UpdateExpression: aws.String("set exp = :dayExp"),
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":dayExp": {N: aws.String(strconv.Itoa(dayEXP))},
					},
					TableName: TableName,
				}
				_, err = svc.UpdateItem(update_input)
				if err != nil {
					fmt.Printf("Update daily XP counter failed: %v\n", err)
					continue
				}

				update_input = &dynamodb.UpdateItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"primary":   {S: aws.String(creds.Username)},
						"secondary": {S: aws.String("profile")},
					},
					UpdateExpression: aws.String("add experience :delta"),
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":delta": {N: aws.String(strconv.Itoa(delta))},
					},
					TableName: TableName,
				}
				_, err = svc.UpdateItem(update_input)
				if err != nil {
					fmt.Printf("Update profile XP failed: %v\n", err)
				}

			}
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
			}, nil
		}
	}
}

func main() {
	lambda.Start(handler)
}
