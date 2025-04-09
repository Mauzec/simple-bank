package token

import (
	"testing"
	"time"

	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
)

func TestPasetoSMaker(t *testing.T) {
	maker, err := NewPasetoSMaker(util.RandomString(32))
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

func TestExpiredPasetoSToken(t *testing.T) {
	maker, err := NewPasetoSMaker(util.RandomString(32))
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
