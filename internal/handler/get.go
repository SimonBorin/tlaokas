package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"tlaokas/internal/crypto"
	"tlaokas/internal/db"
	"tlaokas/internal/model"
)

// GetSecret handles GET /secret/{id} requests to retrieve and decrypt a secret.
func GetSecret(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/secret/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	var s model.Secret
	err = db.DB.QueryRow(context.Background(), `
		SELECT encrypted_data, iv, expires_at, viewed
		FROM secrets
		WHERE id = $1
	`, id).Scan(&s.EncryptedData, &s.IV, &s.ExpiresAt, &s.Viewed)
	if err != nil || s.Viewed || time.Now().After(s.ExpiresAt) {
		http.Error(w, "secret not found or expired", http.StatusNotFound)
		return
	}

	_, err = db.DB.Exec(context.Background(), `UPDATE secrets SET viewed = true WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "failed to mark as viewed", http.StatusInternalServerError)
		return
	}

	plaintext, err := crypto.Decrypt(s.EncryptedData, s.IV)
	if err != nil {
		http.Error(w, "decryption failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, plaintext)
}
