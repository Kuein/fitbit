package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	clientId     = "23BMSG"
	clientSecret = "dabf4bc970ad5d9d709f096da275000b"
	redirect     = "https://5w8t9dch57.execute-api.eu-central-1.amazonaws.com/Prod/webhook"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.Path {
	case "/auth_code":
		{
			// return redirect to fitbit auth page
			url, err := http.NewRequest("GET", "https://www.fitbit.com/oauth2/authorize", nil)
			if err != nil {
				return events.APIGatewayProxyResponse{
					Body:       "Cannot create requests",
					StatusCode: 500,
				}, nil
			}
			q := url.URL.Query()
			q.Add("client_id", clientId)
			q.Add("scope", "weight profile nutrition activity sleep")
			q.Add("redirect_url", redirect)
			q.Add("response_type", "code")
			url.URL.RawQuery = q.Encode()
			return events.APIGatewayProxyResponse{
				StatusCode: 301,
				Headers:    map[string]string{"Location": url.URL.String()},
			}, nil

		}
	default:
		{
			// parse GET url and get CODE from URL
			code := request.QueryStringParameters["code"]
			if len(code) == 0 {
				return events.APIGatewayProxyResponse{
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
				return events.APIGatewayProxyResponse{}, err
			}
			url.Header.Add("Authorization", "Basic "+basic)
			url.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			c := &http.Client{}
			res, err := c.Do(url)
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

			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("Get response:\n%v\n", string(data)),
				StatusCode: 200,
			}, nil

		}
	}
}

func main() {
	lambda.Start(handler)
}