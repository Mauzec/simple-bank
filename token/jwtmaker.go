package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtsMinKeySize = 64

// JWTSMaker is a struct that provides functionality for creating and verifying JWT tokens.
//
// It implements ONLY symmetric key algorithm
type JWTSMaker struct {
	key string
}

func NewJWTSMaker(key string) (Maker, error) {
	if len(key) < jwtsMinKeySize {
		return nil, fmt.Errorf("key size must be at least %v, got %v", jwtsMinKeySize, len(key))
	}

	return &JWTSMaker{key}, nil
}

func (maker *JWTSMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtsToken.SignedString([]byte(maker.key))
}

func (maker *JWTSMaker) keyFunc(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrInvalidToken
	}
	return []byte(maker.key), nil
}

func (maker *JWTSMaker) VerifyToken(token string) (*Payload, error) {
	p := &Payload{}
	parsedToken, err := jwt.ParseWithClaims(token, p, maker.keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}
	if _, ok := parsedToken.Claims.(*Payload); !ok {
		return nil, ErrInvalidToken
	}

	return p, nil
}
