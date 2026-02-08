package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed home.html
var homeHTML []byte

type apiResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Feature string `json:"feature,omitempty"`
	Version string `json:"version,omitempty"`
	Time    string `json:"time,omitempty"`
}

const (
	defaultAppVersion = "local-dev"
	featureProbeName  = "probe"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/feature-probe", featureProbeHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("microservice is listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(homeHTML); err != nil {
		log.Printf("failed to write home page: %v", err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "world"
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Message: "hello, " + name,
	})
}

func featureProbeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Status:  "ok",
		Feature: featureProbeName,
		Version: currentAppVersion(),
	})
}

func currentAppVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		return defaultAppVersion
	}
	return version
}

func writeJSON(w http.ResponseWriter, statusCode int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
