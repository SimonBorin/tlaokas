package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Secret represents the structure stored in the database.
type Secret struct {
	EncryptedData []byte
	IV            []byte
	ExpiresAt     time.Time
	Viewed        bool
}

var (
	db  *pgxpool.Pool
	key = []byte("this_is_32_bytes_secret_key!!!!!") // 32 bytes key for AES-256 encryption
)

// initDB initializes the database connection and creates the "secrets" table if it doesn't exist.
func initDB() error {
	dsn := os.Getenv("DB_DSN")
	var err error
	db, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS secrets (
		id UUID PRIMARY KEY,
		encrypted_data BYTEA NOT NULL,
		iv BYTEA NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		viewed BOOLEAN NOT NULL DEFAULT false
	);`

	_, err = db.Exec(context.Background(), createTable)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// encrypt encrypts a plaintext string using AES-CTR and returns the ciphertext and IV.
func encrypt(plaintext string) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, []byte(plaintext))
	return ciphertext, iv, nil
}

// decrypt decrypts ciphertext using AES-CTR with the given IV.
func decrypt(ciphertext, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)
	return string(plaintext), nil
}

// createSecretHandler handles POST /secret requests to store a new encrypted secret.
func createSecretHandler(w http.ResponseWriter, r *http.Request) {
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

	ciphertext, iv, err := encrypt(req.Secret)
	if err != nil {
		http.Error(w, "encryption failed", http.StatusInternalServerError)
		return
	}

	id := uuid.New()
	expiresAt := time.Now().Add(12 * time.Hour)

	// Insert encrypted secret into the database
	_, err = db.Exec(context.Background(), `
		INSERT INTO secrets (id, encrypted_data, iv, expires_at, viewed)
		VALUES ($1, $2, $3, $4, $5)
	`, id, ciphertext, iv, expiresAt, false)
	if err != nil {
		http.Error(w, "failed to store secret", http.StatusInternalServerError)
		return
	}

	res := response{URL: fmt.Sprintf("http://localhost:8080/secret/%s", id)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// getSecretHandler handles GET /secret/{id} requests to retrieve and decrypt a secret.
func getSecretHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/secret/"):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
	}

	var s Secret
	// Fetch the secret from the database
	err = db.QueryRow(context.Background(), `
		SELECT encrypted_data, iv, expires_at, viewed
		FROM secrets
		WHERE id = $1
	`, id).Scan(&s.EncryptedData, &s.IV, &s.ExpiresAt, &s.Viewed)
	if err != nil || s.Viewed || time.Now().After(s.ExpiresAt) {
		http.Error(w, "secret not found or expired", http.StatusNotFound)
		return
	}

	// Mark the secret as viewed
	_, err = db.Exec(context.Background(), `UPDATE secrets SET viewed = true WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "failed to mark as viewed", http.StatusInternalServerError)
		return
	}

	plaintext, err := decrypt(s.EncryptedData, s.IV)
	if err != nil {
		http.Error(w, "decryption failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, plaintext)
}

// cleanupExpiredSecrets periodically deletes secrets that were viewed or expired.
func cleanupExpiredSecrets() {
	for {
		time.Sleep(10 * time.Minute)
		_, err := db.Exec(context.Background(), `
			DELETE FROM secrets
			WHERE viewed = true OR expires_at < now()
		`)
		if err != nil {
			log.Println("failed to cleanup expired secrets:", err)
		}
	}
}

// withCORS adds CORS headers to the response for cross-origin access.
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			return
		}
		h.ServeHTTP(w, r)
	})
}

// main initializes the database, starts the cleanup goroutine and the HTTP server.
func main() {
	if err := initDB(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	go cleanupExpiredSecrets()

	http.Handle("/secret", withCORS(http.HandlerFunc(createSecretHandler)))
	http.Handle("/secret/", withCORS(http.HandlerFunc(getSecretHandler)))

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
