package handlers

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLoginHandler_BadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Result().StatusCode)
	}
}

func TestLoginHandler_InvalidIDToken(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "client")
	t.Setenv("JWT_SECRET", "secret")

	origClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"sub":"123","email":"user@example.com","aud":"other","exp":"1"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	defer func() { http.DefaultClient = origClient }()

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"idToken":"token"}`))
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Result().StatusCode)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "client")
	t.Setenv("JWT_SECRET", "secret")

	origClient := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"sub":"123","email":"user@example.com","aud":"client","exp":"1"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	defer func() { http.DefaultClient = origClient }()

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"idToken":"token"}`))
	w := httptest.NewRecorder()

	LoginHandler(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("token not returned")
	}

	parts := strings.Split(resp.Token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid token format: %s", resp.Token)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		t.Fatalf("failed to unmarshal claims: %v", err)
	}

	if claims["sub"] != "123" || claims["email"] != "user@example.com" {
		t.Fatalf("unexpected claims: %v", claims)
	}
}
