package token

import "errors"

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token is expired")

	ErrPayloadID = errors.New("unexpected error when creating UUID for payload")
)
