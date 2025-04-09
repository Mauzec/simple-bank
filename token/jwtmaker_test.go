package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
)

func TestJWTSMaker(t *testing.T) {
	maker, err := NewJWTSMaker(util.RandomString(64))
	assert.NoError(t, err)

	username := "fallen_angel"
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := time.Now().Add(duration)

	token, err := maker.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	p, err := maker.VerifyToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotZero(t, p.ID)
	assert.Equal(t, username, p.Username)
	assert.WithinDuration(t, issuedAt, p.IssuedAt, time.Second)
	assert.WithinDuration(t, expiredAt, p.ExpiredAt, time.Second)
}

func TestExpiredJWTSToken(t *testing.T) {
	maker, err := NewJWTSMaker(util.RandomString(64))
	assert.NoError(t, err)

	username := "fallen_angel"
	duration := time.Nanosecond

	token, err := maker.CreateToken(username, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	p, err := maker.VerifyToken(token)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrExpiredToken)
	assert.Nil(t, p)
}

func TestInvalidJWTSTokenAlgNone(t *testing.T) {
	p, err := NewPayload(util.RandomOwner(), time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, p)

	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, p).SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	maker, err := NewJWTSMaker(util.RandomString(64))
	assert.NoError(t, err)

	p1, err := maker.VerifyToken(token)
	assert.Nil(t, p1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
