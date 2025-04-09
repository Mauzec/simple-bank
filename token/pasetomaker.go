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
	p, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	token, err := maker.paseto.Encrypt(maker.key, p, "Mauzec")
	return token, err
}

func (maker *PasetoSMaker) VerifyToken(token string) (*Payload, error) {
	p := &Payload{}

	var footer string
	err := maker.paseto.Decrypt(token, maker.key, p, &footer)
	if err != nil || footer != "Mauzec" {
		return nil, ErrInvalidToken
	}
	err = p.Valid()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewPasetoSMaker(key string) (*PasetoSMaker, error) {
	if len(key) < chacha20poly1305.KeySize {
		return nil, fmt.Errorf("key size must be equal to %v, got %v", chacha20poly1305.KeySize, len(key))
	}
	return &PasetoSMaker{paseto.NewV2(), []byte(key)}, nil
}
