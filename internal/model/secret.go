package model

import "time"

// Secret represents the structure stored in the database.
type Secret struct {
	EncryptedData []byte
	IV            []byte
	ExpiresAt     time.Time
	Viewed        bool
}
