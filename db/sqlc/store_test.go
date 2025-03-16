package db

import (
	"context"
	"testing"

	"github.com/mauzec/simple-bank/db/util"
	"github.com/stretchr/testify/assert"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createAndTestRandomAccount(t)
	account2 := createAndTestRandomAccount(t)
	amountFloat := float64(15)
	amount, err := util.Float64ToNumeric(amountFloat)
	assert.NoError(t, err)

	// let check concurrency transfering

	errs := make(chan error)
	results := make(chan TransferTxResult)

	n := 5
	for range n {
		go func() {
			// fmt.Printf("%d: Start transfer transaction\n", i)
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			// fmt.Printf("%d: End transfer transaction\n", i)

			errs <- err
			results <- result
		}()
	}

	xs := make([]bool, n)

	for range n {
		err := <-errs
		assert.NoError(t, err)

		result := <-results
		assert.NotEmpty(t, result)

		// check transfers

		transfer := result.Transfer
		assert.NotEmpty(t, transfer)
		assert.NotZero(t, transfer.ID)
		assert.Equal(t, account1.ID, transfer.FromAccountID)
		assert.Equal(t, account2.ID, transfer.ToAccountID)
		assert.Equal(t, amount, transfer.Amount)
		assert.NotZero(t, transfer.CreatedAt)

		// check entries

		fromEntry := result.FromEntry
		assert.NotEmpty(t, result.FromEntry)
		assert.NotZero(t, fromEntry.ID)
		assert.Equal(t, account1.ID, fromEntry.AccountID)
		assert.NotZero(t, fromEntry.CreatedAt)

		toEntry := result.ToEntry
		assert.NotEmpty(t, result.ToEntry)
		assert.NotZero(t, toEntry.ID)
		assert.Equal(t, account2.ID, toEntry.AccountID)
		assert.NotZero(t, toEntry.CreatedAt)

		getFromEntry, err := store.GetEntry(context.Background(), fromEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, getFromEntry, fromEntry)

		getToEntry, err := store.GetEntry(context.Background(), toEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, getToEntry, toEntry)

		// check that accounts' data chancged

		fromAccount := result.FromAccount
		toAccount := result.ToAccount
		assert.NotEmpty(t, fromAccount)
		assert.NotEmpty(t, toAccount)

		_, d1, err := util.AddNumToNum(&account1.Balance, &fromAccount.Balance, -1)
		assert.NoError(t, err)
		_, d2, err := util.AddNumToNum(&toAccount.Balance, &account2.Balance, -1)
		assert.NoError(t, err)
		assert.True(t, d1 > 0)
		assert.Equal(t, d1, d2)

		// want: equal(t, i * amount, i * amount), where i is number of accounts'
		// updates within concurrent running transaction
		assert.True(t, util.IsInt(d1/amountFloat))

		// to ensure that all goroutines do their job right
		xnometer := int(d1 / amountFloat)
		assert.True(t, xnometer >= 1 && xnometer <= n)
		assert.Equal(t, false, xs[xnometer-1])
		xs[xnometer] = true
	}

	updAcc1, err := store.GetAccount(context.Background(), account1.ID)
	assert.NoError(t, err)
	updAcc2, err := store.GetAccount(context.Background(), account2.ID)
	assert.NoError(t, err)

	expectAcc1Balance, _, err := util.AddNumToNum(account1.Balance, float64(n)*amountFloat, -1)
	assert.NoError(t, err)
	expectAcc2Balance, _, err := util.AddNumToNum(account2.Balance, float64(n)*amountFloat)
	assert.NoError(t, err)

	assert.Equal(t, expectAcc1Balance, updAcc1.Balance)
	assert.Equal(t, expectAcc2Balance, updAcc2.Balance)
}
