package cleanup

import (
	"context"
	"log"
	"time"

	"tlaokas/internal/db"
)

// Run starts a background job that periodically deletes viewed or expired secrets.
func Run() {
	for {
		time.Sleep(10 * time.Minute)
		_, err := db.DB.Exec(context.Background(), `
			DELETE FROM secrets
			WHERE viewed = true OR expires_at < now()
		`)
		if err != nil {
			log.Println("failed to cleanup expired secrets:", err)
		}
	}
}
