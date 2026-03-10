package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Health(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/health" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	}))
	defer srv.Close()

	c := New(srv.URL)
	resp, err := c.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("Status = %q, want %q", resp.Status, "ok")
	}
}

func TestClient_ListAgents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/agents" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ListAgentsResponse{TotalCount: 2})
	}))
	defer srv.Close()

	c := New(srv.URL)
	resp, err := c.ListAgents(context.Background(), ListAgentsOptions{})
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestClient_RegisterAndDelete(t *testing.T) {
	var registered bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agents":
			registered = true
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"id":   "test-id",
				"name": "TestAgent",
			})
		case r.Method == http.MethodDelete:
			if !registered {
				t.Error("delete before register")
			}
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()

	c := New(srv.URL)
	ctx := context.Background()

	card, err := c.RegisterAgent(ctx, RegisterRequest{
		Name:      "TestAgent",
		Protocols: []string{"a2a"},
		Endpoint:  EndpointReq{URL: "http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("RegisterAgent() error = %v", err)
	}
	if card.Name != "TestAgent" {
		t.Errorf("Name = %q, want TestAgent", card.Name)
	}

	if err := c.DeleteAgent(ctx, "test-id"); err != nil {
		t.Fatalf("DeleteAgent() error = %v", err)
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
	}))
	defer srv.Close()

	c := New(srv.URL)
	_, err := c.Health(context.Background())
	if err == nil {
		t.Error("expected error for 400 response")
	}
}
