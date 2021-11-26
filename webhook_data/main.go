package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println(request)
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
	case "POST":
		{
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
	}, nil
}

func main() {
	lambda.Start(handler)
}
