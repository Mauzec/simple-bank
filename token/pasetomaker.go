package token

import (
	"fmt"

	"time"

	"github.com/o1egl/paseto/v2"
	"golang.org/x/crypto/chacha20poly1305"
)

type PasetoSMaker struct {
	paseto *paseto.V2
	key    []byte
}

func (maker *PasetoSMaker) CreateToken(username string, duration time.Duration) (string, error) {
	panic("TODO: Implement")
}

func (maker *PasetoSMaker) VerifyToken(token string) (*Payload, error) {
	panic("TODO: Implement")
}

func NewPasetoSMaker(key string) (*PasetoSMaker, error) {
	if len(key) < chacha20poly1305.KeySize {
		return nil, fmt.Errorf("key size must be at least %v, got %v", chacha20poly1305.KeySize, len(key))
	}
	return &PasetoSMaker{paseto.NewV2(), []byte(key)}, nil
}
