package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

//go:embed home.html
var homeHTML string

type apiResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Time    string `json:"time,omitempty"`
}

func main() {
	lambda.Start(handle)
}

func handle(_ context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	path := req.RawPath
	if path == "" {
		path = "/"
	}

	method := strings.ToUpper(req.RequestContext.HTTP.Method)

	switch path {
	case "/":
		if method != http.MethodGet {
			return textResponse(http.StatusMethodNotAllowed, "method not allowed"), nil
		}
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "text/html; charset=utf-8",
			},
			Body: homeHTML,
		}, nil
	case "/healthz":
		if method != http.MethodGet {
			return textResponse(http.StatusMethodNotAllowed, "method not allowed"), nil
		}

		body, err := json.Marshal(apiResponse{
			Status: "ok",
			Time:   time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			return textResponse(http.StatusInternalServerError, "encoding error"), nil
		}

		return jsonResponse(http.StatusOK, string(body)), nil
	case "/hello":
		if method != http.MethodGet {
			return textResponse(http.StatusMethodNotAllowed, "method not allowed"), nil
		}

		name := "world"
		if raw := req.RawQueryString; raw != "" {
			values, err := url.ParseQuery(raw)
			if err == nil {
				if v := values.Get("name"); v != "" {
					name = v
				}
			}
		}

		body, err := json.Marshal(apiResponse{
			Message: "hello, " + name,
		})
		if err != nil {
			return textResponse(http.StatusInternalServerError, "encoding error"), nil
		}

		return jsonResponse(http.StatusOK, string(body)), nil
	default:
		return textResponse(http.StatusNotFound, "not found"), nil
	}
}

func jsonResponse(statusCode int, body string) events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
}

func textResponse(statusCode int, body string) events.LambdaFunctionURLResponse {
	return events.LambdaFunctionURLResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
		},
		Body: body,
	}
}
