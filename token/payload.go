package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Payload represents the structure of the part of data contained in a token.
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: p.ExpiredAt}, nil
}

func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return &jwt.NumericDate{Time: p.IssuedAt}, nil
}

func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	// need to change?
	return &jwt.NumericDate{Time: p.IssuedAt}, nil
}

func (p *Payload) GetIssuer() (string, error) {
	return "fill_me", nil
}

func (p *Payload) GetSubject() (string, error) {
	return p.ID.String(), nil
}

func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, ErrPayloadID
	}

	return &Payload{
		id,
		username,
		time.Now(),
		time.Now().Add(duration),
	}, nil
}
