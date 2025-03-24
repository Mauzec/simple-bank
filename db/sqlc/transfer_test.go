package db

import (
	"context"
	"testing"

	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
)

func createAndTestTransfer(t *testing.T, account1 *Account, account2 *Account) Transfer {
	args := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomBalance(),
	}

	transfer, err := testQueries.CreateTransfer(
		context.Background(), args,
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, transfer)

	assert.NotZero(t, transfer.ID)
	assert.NotZero(t, transfer.CreatedAt)

	assert.Equal(t, account1.ID, transfer.FromAccountID)
	assert.Equal(t, account2.ID, transfer.ToAccountID)
	assert.Equal(t, args.Amount, transfer.Amount)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	account1 := createAndTestRandomAccount(t)
	account2 := createAndTestRandomAccount(t)
	createAndTestTransfer(t, &account1, &account2)
}

func TestGetTransfer(t *testing.T) {
	account1 := createAndTestRandomAccount(t)
	account2 := createAndTestRandomAccount(t)
	transfer := createAndTestTransfer(t, &account1, &account2)

	actualTransfer, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, actualTransfer)

	assert.Equal(t, transfer.ID, actualTransfer.ID)
	assert.Equal(t, transfer.Amount, actualTransfer.Amount)
	assert.Equal(t, transfer.FromAccountID, actualTransfer.FromAccountID)
	assert.Equal(t, transfer.ToAccountID, actualTransfer.ToAccountID)
	assert.Equal(t, transfer.CreatedAt, actualTransfer.CreatedAt)
}

func TestListTransfers(t *testing.T) {
	// test only after clearing `transfers` table

	account1 := createAndTestRandomAccount(t)
	account2 := createAndTestRandomAccount(t)
	transfers := make([]Transfer, 100)
	for i := range 100 {
		transfers[i] = createAndTestTransfer(t, &account1, &account2)
	}

	args := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Limit:         100,
		Offset:        12,
	}

	actualTransfers, err := testQueries.ListTransfers(context.Background(), args)
	assert.NoError(t, err)
	assert.Len(t, actualTransfers, 88)

	for i, actualTransfer := range actualTransfers {
		assert.NotEmpty(t, actualTransfer)

		transfer := transfers[i+12]
		assert.Equal(t, transfer.ID, actualTransfer.ID)
		assert.Equal(t, transfer.Amount, actualTransfer.Amount)
		assert.Equal(t, transfer.FromAccountID, actualTransfer.FromAccountID)
		assert.Equal(t, transfer.ToAccountID, actualTransfer.ToAccountID)
		assert.Equal(t, transfer.CreatedAt, actualTransfer.CreatedAt)
	}
}
