package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/feel-coding/tessera-data/internal/handler"
	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
)

// makeEmbedding returns a 1536-float32 slice where every element is val.
func makeEmbedding(val float32) []float32 {
	e := make([]float32, 1536)
	for i := range e {
		e[i] = val
	}
	return e
}

// TestSaveAndFindSimilarCase inserts a case via POST /cases then retrieves it via
// POST /cases/similar with the same embedding and asserts similarity > 0.99.
func TestSaveAndFindSimilarCase(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	casesH := handler.NewCases(pool)
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Post("/cases", casesH.Save)
	r.Post("/cases/similar", casesH.FindSimilar)
	server := httptest.NewServer(r)
	defer server.Close()

	embedding := makeEmbedding(0.5)

	saveBody := service.SaveCaseInput{
		TransactionID: "txn_cases_test_001",
		Decision:      "APPROVE",
		Reasoning:     "Integration test case — known device, low amount.",
		Signals:       map[string]any{"blacklist_hit": false},
		Embedding:     embedding,
		Metadata:      map[string]any{"source": "test"},
	}
	saveBytes, _ := json.Marshal(saveBody)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/cases", bytes.NewReader(saveBytes))
	req.Header.Set("X-Internal-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /cases request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 from POST /cases, got %d", resp.StatusCode)
	}

	var saveResp struct {
		ID            string `json:"id"`
		TransactionID string `json:"transaction_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&saveResp); err != nil {
		t.Fatalf("failed to decode save response: %v", err)
	}
	if saveResp.ID == "" {
		t.Fatal("expected non-empty id in save response")
	}

	searchBody := service.SimilarSearchInput{
		Embedding: embedding,
		Limit:     5,
	}
	searchBytes, _ := json.Marshal(searchBody)

	req2, _ := http.NewRequest(http.MethodPost, server.URL+"/cases/similar", bytes.NewReader(searchBytes))
	req2.Header.Set("X-Internal-Key", "test-key")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("POST /cases/similar request failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from POST /cases/similar, got %d", resp2.StatusCode)
	}

	var searchResp struct {
		Cases []struct {
			ID         string  `json:"id"`
			Similarity float64 `json:"similarity"`
		} `json:"cases"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&searchResp); err != nil {
		t.Fatalf("failed to decode search response: %v", err)
	}
	if len(searchResp.Cases) == 0 {
		t.Fatal("expected at least one case in similarity search result")
	}

	found := false
	for _, c := range searchResp.Cases {
		if c.ID == saveResp.ID {
			found = true
			if c.Similarity < 0.99 {
				t.Errorf("expected similarity >= 0.99 for exact embedding match, got %f", c.Similarity)
			}
			break
		}
	}
	if !found {
		t.Errorf("inserted case %s not found in similarity search results", saveResp.ID)
	}
}

// TestFindSimilarInvalidEmbeddingLength verifies that a wrong-length embedding returns 400.
func TestFindSimilarInvalidEmbeddingLength(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	casesH := handler.NewCases(pool)
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Post("/cases/similar", casesH.FindSimilar)
	server := httptest.NewServer(r)
	defer server.Close()

	body := service.SimilarSearchInput{
		Embedding: []float32{0.1, 0.2, 0.3}, // wrong length
		Limit:     5,
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/cases/similar", bytes.NewReader(b))
	req.Header.Set("X-Internal-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// TestSaveCaseMissingFields verifies that missing transaction_id or decision returns 400.
func TestSaveCaseMissingFields(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	casesH := handler.NewCases(pool)
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Post("/cases", casesH.Save)
	server := httptest.NewServer(r)
	defer server.Close()

	body := service.SaveCaseInput{
		TransactionID: "", // intentionally empty
		Decision:      "",
		Embedding:     makeEmbedding(0.1),
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/cases", bytes.NewReader(b))
	req.Header.Set("X-Internal-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
