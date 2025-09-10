package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"tlaokas/internal/crypto"
	"tlaokas/internal/db"
)

// CreateSecret handles POST /secret requests to store a new encrypted secret.
func CreateSecret(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Secret string `json:"secret"`
	}
	type response struct {
		URL string `json:"url"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ciphertext, iv, err := crypto.Encrypt(req.Secret)
	if err != nil {
		http.Error(w, "encryption failed", http.StatusInternalServerError)
		return
	}

	id := uuid.New()
	expiresAt := time.Now().Add(12 * time.Hour)

	_, err = db.DB.Exec(context.Background(), `
		INSERT INTO secrets (id, encrypted_data, iv, expires_at, viewed)
		VALUES ($1, $2, $3, $4, $5)
	`, id, ciphertext, iv, expiresAt, false)
	if err != nil {
		http.Error(w, "failed to store secret", http.StatusInternalServerError)
		return
	}

	res := response{URL: fmt.Sprintf("/secret/%s", id)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
