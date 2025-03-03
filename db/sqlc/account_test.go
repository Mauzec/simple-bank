package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mauzec/simple-bank/db/util"
	"github.com/stretchr/testify/assert"
)

const (
	dbSource = "postgresql://root:secret@localhost:5431/simple_bank?sslmode=disable"
)

var testQueries *Queries

func createAndTestRandomAccount(t *testing.T) Account {
	args := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}

	ctx := context.Background()
	account, err := testQueries.CreateAccount(ctx, args)
	if !assert.NoError(t, err) {
		log.Fatal(
			fmt.Errorf(
				"Unable to create account with arguments:\n %+v \n %v", args, err),
		)
	}
	if !assert.NotEmpty(t, account) {
		log.Fatal(
			fmt.Errorf(
				"Received empty account, but arguments:\n %+v", args),
		)
	}

	assert.Equal(t, args.Balance, account.Balance)
	assert.Equal(t, args.Owner, account.Owner)
	assert.Equal(t, args.Currency, account.Currency)

	assert.NotZero(t, account.ID)
	assert.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createAndTestRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	var account Account

	t.Run("Creating Account", func(t *testing.T) {
		account = createAndTestRandomAccount(t)
	})

	t.Run("Getting Account", func(t *testing.T) {
		ctx := context.Background()
		actualAccount, err := testQueries.GetAccount(ctx, account.ID)
		if !assert.NoError(t, err) {
			log.Fatal(
				fmt.Errorf(
					"Unable to get account with id=%d \n %+v ", account.ID, err),
			)
		}
		if !assert.NotEmpty(t, actualAccount) {
			log.Fatal(
				fmt.Errorf(
					"Received empty account, id=%d", account.ID),
			)
		}

		assert.Equal(t, account.Balance, actualAccount.Balance)
		assert.Equal(t, account.Owner, actualAccount.Owner)
		assert.Equal(t, account.Currency, actualAccount.Currency)
		assert.WithinDuration(t, account.CreatedAt.Time, actualAccount.CreatedAt.Time, time.Second)
		assert.Equal(t, account.ID, actualAccount.ID)
	})
}

func TestUpdateAccount(t *testing.T) {
	var account Account

	t.Run("Creating Account", func(t *testing.T) {
		account = createAndTestRandomAccount(t)
	})

	t.Run("Updating Account", func(t *testing.T) {
		wantArgs := UpdateAccountParams{
			ID:      account.ID,
			Balance: 2077,
		}

		newAccount, err := testQueries.UpdateAccount(context.Background(), wantArgs)
		if !assert.NoError(t, err) {
			log.Fatal(
				fmt.Errorf(
					"Unable to update account with id=%d, newBalance=%d \n %+v ", wantArgs.ID, wantArgs.Balance, err),
			)
		}
		if !assert.NotEmpty(t, newAccount) {
			log.Fatal(
				fmt.Errorf(
					"Received empty account, id=%d", wantArgs.ID),
			)
		}

		assert.Equal(t, wantArgs.Balance, newAccount.Balance)
		assert.Equal(t, account.Owner, newAccount.Owner)
		assert.Equal(t, account.Currency, newAccount.Currency)
		assert.Equal(t, account.CreatedAt, newAccount.CreatedAt)
		assert.Equal(t, wantArgs.ID, newAccount.ID)
	})
}

func TestDeleteAccount(t *testing.T) {
	var account Account
	t.Run("Creating Account", func(t *testing.T) {
		account = createAndTestRandomAccount(t)
	})

	t.Run("Deleting Account", func(t *testing.T) {
		err := testQueries.DeleteAccount(context.Background(), account.ID)
		if !assert.NoError(t, err) {
			log.Fatal(
				fmt.Errorf(
					"Unable to delete account=%+v,\n %v \n", account, err),
			)
		}
	})

	t.Run("Getting Account", func(t *testing.T) {
		actAccount, err := testQueries.GetAccount(context.Background(), account.ID)
		assert.Error(t, err)
		assert.EqualError(t, err, pgx.ErrNoRows.Error())
		assert.Empty(t, actAccount)
	})
}

func TestListAccounts(t *testing.T) {
	for range 10 {
		createAndTestRandomAccount(t)
	}

	args := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), args)
	assert.NoError(t, err)
	assert.NotEmpty(t, accounts)
	assert.Len(t, accounts, 5)

	for _, acc := range accounts {
		assert.NotEmpty(t, acc)
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbSource)
	if err != nil {
		log.Fatal("unable to connect to database", err)
	}
	defer conn.Close(ctx)

	testQueries = New(conn)

	os.Exit(m.Run())
}
