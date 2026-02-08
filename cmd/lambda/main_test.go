package main

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleHome(t *testing.T) {
	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if !strings.Contains(res.Headers["Content-Type"], "text/html") {
		t.Fatalf("expected html content type, got %q", res.Headers["Content-Type"])
	}

	if !strings.Contains(res.Body, "Service dashboard") {
		t.Fatalf("expected dashboard html body, got %q", res.Body)
	}
}

func TestHandleHealthz(t *testing.T) {
	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/healthz",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if !strings.Contains(res.Body, "\"status\":\"ok\"") {
		t.Fatalf("expected ok json body, got %q", res.Body)
	}
}

func TestHandleHello(t *testing.T) {
	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath:        "/hello",
		RawQueryString: "name=AWS",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if !strings.Contains(res.Body, "\"message\":\"hello, AWS\"") {
		t.Fatalf("expected hello json body, got %q", res.Body)
	}
}

func TestHandleNotFound(t *testing.T) {
	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/not-found",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestHandleMethodNotAllowed(t *testing.T) {
	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/hello",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodPost,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, res.StatusCode)
	}
}

func TestHandleFeatureProbe(t *testing.T) {
	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/feature-probe",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if !strings.Contains(res.Body, "\"status\":\"ok\"") {
		t.Fatalf("expected probe status, got %q", res.Body)
	}

	if !strings.Contains(res.Body, "\"feature\":\""+featureProbeName+"\"") {
		t.Fatalf("expected probe feature name, got %q", res.Body)
	}

	if !strings.Contains(res.Body, "\"version\":\""+defaultAppVersion+"\"") {
		t.Fatalf("expected default app version, got %q", res.Body)
	}
}

func TestHandleFeatureProbeVersionFromEnv(t *testing.T) {
	t.Setenv("APP_VERSION", "sha-test")

	res, err := handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/feature-probe",
		RequestContext: events.LambdaFunctionURLRequestContext{
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(res.Body, "\"version\":\"sha-test\"") {
		t.Fatalf("expected version from env, got %q", res.Body)
	}
}
