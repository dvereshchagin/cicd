package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected Content-Type text/html, got %q", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Service dashboard") {
		t.Fatalf("expected home page body, got %q", body)
	}
}

func TestHomeHandlerNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestHomeHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	homeHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	healthzHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Status != "ok" {
		t.Fatalf("expected status 'ok', got %q", payload.Status)
	}
}

func TestHelloHandlerDefaultName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()

	helloHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Message != "hello, world" {
		t.Fatalf("expected message %q, got %q", "hello, world", payload.Message)
	}
}

func TestHelloHandlerWithName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hello?name=dev", nil)
	rec := httptest.NewRecorder()

	helloHandler(rec, req)

	var payload apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Message != "hello, dev" {
		t.Fatalf("expected message %q, got %q", "hello, dev", payload.Message)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	rec := httptest.NewRecorder()

	healthzHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestFeatureProbeHandlerDefaultVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/feature-probe", nil)
	rec := httptest.NewRecorder()

	featureProbeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var payload apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Status != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", payload.Status)
	}

	if payload.Feature != featureProbeName {
		t.Fatalf("expected feature %q, got %q", featureProbeName, payload.Feature)
	}

	if payload.Version != defaultAppVersion {
		t.Fatalf("expected version %q, got %q", defaultAppVersion, payload.Version)
	}
}

func TestFeatureProbeHandlerFromEnv(t *testing.T) {
	t.Setenv("APP_VERSION", "sha-test")

	req := httptest.NewRequest(http.MethodGet, "/feature-probe", nil)
	rec := httptest.NewRecorder()

	featureProbeHandler(rec, req)

	var payload apiResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.Version != "sha-test" {
		t.Fatalf("expected version %q, got %q", "sha-test", payload.Version)
	}
}
