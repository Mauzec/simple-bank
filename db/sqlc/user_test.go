package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(10))
	assert.NoError(t, err)
	args := CreateUserParams{
		Username:       util.RandomOwner(),
		FullName:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		Email:          util.RandomEmail(),
	}

	ctx := context.Background()
	user, err := testQueries.CreateUser(ctx, args)
	if !assert.NoError(t, err) {
		log.Fatal(
			fmt.Errorf(
				"Unable to create user with arguments:\n %+v \n %v", args, err),
		)
	}
	if !assert.NotEmpty(t, user) {
		log.Fatal(
			fmt.Errorf(
				"Received empty user, but arguments:\n %+v", args),
		)
	}

	assert.Equal(t, args.Username, user.Username)
	assert.Equal(t, args.FullName, user.FullName)
	assert.Equal(t, args.HashedPassword, user.HashedPassword)
	assert.Equal(t, args.Email, user.Email)

	assert.True(t, user.PasswordChangedAt.Time.IsZero())
	assert.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	_ = createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)

	gotUser, err := testQueries.GetUser(context.Background(), user.Username)
	assert.NoError(t, err)
	assert.Equal(t, user.FullName, gotUser.FullName)
	assert.Equal(t, user.Username, gotUser.Username)
	assert.Equal(t, user.HashedPassword, gotUser.HashedPassword)
	assert.Equal(t, user.Email, gotUser.Email)
	assert.Equal(t, user.PasswordChangedAt, gotUser.PasswordChangedAt)
	assert.Equal(t, user.CreatedAt, gotUser.CreatedAt)
}
