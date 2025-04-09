package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashingPassword(t *testing.T) {
	password := RandomString(10)

	hashedPassword, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	t.Run("RightPassword", func(t *testing.T) {
		err = CheckPassword(hashedPassword, password)
		assert.NoError(t, err)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		err = CheckPassword(hashedPassword, "hello")
		assert.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())
	})

	t.Run("OtherHashedPassword", func(t *testing.T) {
		hashedPassword2, err := HashPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword2)
		assert.NotEqual(t, hashedPassword, hashedPassword2)
	})
}
