package token

import "time"

// Maker is an interface that defines the methods required for creating and
// verifying tokens.
type Maker interface {
	CreateToken(username string, duration time.Duration) (string, error)
	VerifyToken(token string) (*Payload, error)
}
