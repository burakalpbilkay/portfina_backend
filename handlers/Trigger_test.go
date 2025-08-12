package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestTriggerHandlerRequiresDagID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/trigger/", nil)
	rr := httptest.NewRecorder()

	TriggerHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestTriggerHandlerSuccess(t *testing.T) {
	original := dagTrigger
	dagTrigger = func(string) bool { return true }
	defer func() { dagTrigger = original }()

	req := httptest.NewRequest(http.MethodPost, "/trigger/test", nil)
	req = mux.SetURLVars(req, map[string]string{"dag": "test"})
	rr := httptest.NewRecorder()

	TriggerHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["message"] == "" {
		t.Fatalf("expected message in response")
	}
}

func TestTriggerHandlerFailure(t *testing.T) {
	original := dagTrigger
	dagTrigger = func(string) bool { return false }
	defer func() { dagTrigger = original }()

	req := httptest.NewRequest(http.MethodPost, "/trigger/test", nil)
	req = mux.SetURLVars(req, map[string]string{"dag": "test"})
	rr := httptest.NewRecorder()

	TriggerHandler(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rr.Code)
	}
}
