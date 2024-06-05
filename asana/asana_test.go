package asanaexporter

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	// I tried to use wiremock with test containers but couldn't get to run it locally due to some errors
	// . "github.com/wiremock/wiremock-testcontainers-go"
)

func TestExporter(t *testing.T) {
	successResponse := AsanaResponse{
		Data: []ObjectSchema{
			{
				Gid:          "1",
				Name:         "1",
				ResourceType: "project",
			},
		},
	}
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(successResponse)

	}))
	defer testServer.Close()

	exporter := NewAsanaExtractor(testServer.URL, "")
	r, err := exporter.RetrieveEntities("project")
	if err != nil {
		t.Error("Failed with", err)
	}
	if len(r.Data) == 0 {
		t.Error("No entities found while they should be present")
	}
}

func TestRetryInvalid(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer testServer.Close()

	exporter := NewAsanaExtractor(testServer.URL, "")
	_, err := exporter.RetrieveEntities("project")
	if err == nil {
		t.Error("Retry policy was invalid, request should have failed")
	}
	if !errors.Is(err, ErrFailedRetry) {
		t.Error("Retry policy was invalid, request should have failed with a different error, actual error: ", err)
	}
}

func TestRateLimit(t *testing.T) {
	successResponse := AsanaResponse{
		Data: []ObjectSchema{
			{
				Gid:          "1",
				Name:         "1",
				ResourceType: "project",
			},
		},
	}
	retried := false
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !retried {
			w.Header().Set("content-type", "application/json")
			w.Header().Set("retry-after", "3")
			w.WriteHeader(http.StatusTooManyRequests)
			retried = true
		} else {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(successResponse)
		}
	}))
	defer testServer.Close()

	exporter := NewAsanaExtractor(testServer.URL, "")
	r, err := exporter.RetrieveEntities("project")
	if err != nil {
		t.Error("Failed with", err)
	}
	if len(r.Data) == 0 {
		t.Error("No entities found while they should be present")
	}
}
