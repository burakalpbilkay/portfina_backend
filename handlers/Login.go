package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type loginRequest struct {
	IDToken string `json:"idToken"`
}

type googleTokenInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Aud   string `json:"aud"`
	Exp   string `json:"exp"`
}

func verifyGoogleIDToken(idToken string) (*googleTokenInfo, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid id token")
	}
	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	if info.Aud != os.Getenv("GOOGLE_CLIENT_ID") {
		return nil, fmt.Errorf("invalid audience")
	}
	return &info, nil
}

func signJWT(claims map[string]interface{}, secret []byte) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	msg := header + "." + payload
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msg))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return msg + "." + signature, nil
}

// LoginHandler verifies a Google ID token and returns a signed JWT for the client.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	info, err := verifyGoogleIDToken(req.IDToken)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid id token"})
		return
	}

	claims := map[string]interface{}{
		"sub":   info.Sub,
		"email": info.Email,
		"exp":   time.Now().Add(time.Hour).Unix(),
	}

	tokenStr, err := signJWT(claims, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to sign token"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": tokenStr})
}
