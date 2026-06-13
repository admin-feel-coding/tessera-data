package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/feel-coding/tessera-data/internal/handler"
	"github.com/feel-coding/tessera-data/internal/httpx"
	"github.com/feel-coding/tessera-data/internal/service"
	"github.com/feel-coding/tessera-data/internal/store"
)

var verdictCounter int64

// uniqueID returns a unique integer for generating distinct transaction IDs per test run.
func uniqueID() int64 {
	return atomic.AddInt64(&verdictCounter, 1)
}

// newVerdictsRouter builds a chi router with all three verdict routes.
func newVerdictsRouter(verdictsH *handler.VerdictsHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.AuthMiddleware("test-key"))
	r.Post("/verdicts", verdictsH.Save)
	r.Get("/verdicts", verdictsH.List)
	r.Get("/verdicts/{transaction_id}", verdictsH.GetByTransactionID)
	return r
}

// sampleVerdictInput returns a valid SaveVerdictInput with a unique transaction_id.
func sampleVerdictInput(txnID string) service.SaveVerdictInput {
	return service.SaveVerdictInput{
		TransactionID:   txnID,
		Decision:        "APPROVE",
		RiskScore:       0.12,
		Reasoning:       "Known device, low amount, no blacklist hit.",
		Signals:         map[string]any{"blacklist_hit": false, "device_known": true},
		CitedSources:    []any{map[string]any{"type": "case", "id": "case_001"}},
		LatencyMs:       1234,
		Model:           "claude-sonnet-4-6",
		ToolCalls:       3,
		LangfuseTraceID: "trace_test_001",
	}
}

// TestSaveVerdict verifies that POST /verdicts returns 201 with a non-empty id.
func TestSaveVerdict(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	verdictsH := handler.NewVerdicts(pool)
	server := httptest.NewServer(newVerdictsRouter(verdictsH))
	defer server.Close()

	body := sampleVerdictInput(fmt.Sprintf("txn_verdict_save_%d", uniqueID()))
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/verdicts", bytes.NewReader(b))
	req.Header.Set("X-Internal-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /verdicts request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var saveResp struct {
		ID            string `json:"id"`
		TransactionID string `json:"transaction_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&saveResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if saveResp.ID == "" {
		t.Fatal("expected non-empty id in response")
	}
	if saveResp.TransactionID != body.TransactionID {
		t.Errorf("transaction_id mismatch: got %q, want %q", saveResp.TransactionID, body.TransactionID)
	}
}

// TestListVerdicts verifies that GET /verdicts returns 200 with verdicts and total fields.
func TestListVerdicts(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	verdictsH := handler.NewVerdicts(pool)
	server := httptest.NewServer(newVerdictsRouter(verdictsH))
	defer server.Close()

	// Insert one verdict so the list is non-trivially testable.
	saveBody := sampleVerdictInput(fmt.Sprintf("txn_verdict_list_%d", uniqueID()))
	b, _ := json.Marshal(saveBody)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/verdicts", bytes.NewReader(b))
	req.Header.Set("X-Internal-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("setup POST /verdicts failed: %v", err)
	}

	req2, _ := http.NewRequest(http.MethodGet, server.URL+"/verdicts?limit=10&offset=0", nil)
	req2.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /verdicts request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var listResp struct {
		Verdicts []store.Verdict `json:"verdicts"`
		Total    int             `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if listResp.Total < 1 {
		t.Errorf("expected total >= 1, got %d", listResp.Total)
	}
}

// TestGetVerdictByTransactionID saves a verdict then retrieves it by transaction_id.
func TestGetVerdictByTransactionID(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	verdictsH := handler.NewVerdicts(pool)
	server := httptest.NewServer(newVerdictsRouter(verdictsH))
	defer server.Close()

	txnID := fmt.Sprintf("txn_verdict_get_%d", uniqueID())
	body := sampleVerdictInput(txnID)
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/verdicts", bytes.NewReader(b))
	req.Header.Set("X-Internal-Key", "test-key")
	req.Header.Set("Content-Type", "application/json")
	if _, err := http.DefaultClient.Do(req); err != nil {
		t.Fatalf("setup POST /verdicts failed: %v", err)
	}

	req2, _ := http.NewRequest(http.MethodGet, server.URL+"/verdicts/"+txnID, nil)
	req2.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /verdicts/{transaction_id} request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var v store.Verdict
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if v.TransactionID != txnID {
		t.Errorf("transaction_id mismatch: got %q, want %q", v.TransactionID, txnID)
	}
	if v.Decision != "APPROVE" {
		t.Errorf("decision mismatch: got %q, want APPROVE", v.Decision)
	}
}

// TestGetVerdictNotFound verifies that a non-existent transaction_id returns 404.
func TestGetVerdictNotFound(t *testing.T) {
	pool := connectTestDB(t)
	defer pool.Close()

	verdictsH := handler.NewVerdicts(pool)
	server := httptest.NewServer(newVerdictsRouter(verdictsH))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/verdicts/txn_does_not_exist", nil)
	req.Header.Set("X-Internal-Key", "test-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /verdicts/{transaction_id} request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code NOT_FOUND, got %q", errResp.Error.Code)
	}
}
