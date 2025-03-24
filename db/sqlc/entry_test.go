package db

import (
	"context"
	"testing"

	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
)

func createAndTestRandomEntry(t *testing.T, account *Account) Entry {
	args := CreateEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomBalance(),
	}

	entry, err := testQueries.CreateEntry(
		context.Background(), args,
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, entry)

	assert.NotZero(t, entry.ID)
	assert.NotZero(t, entry.CreatedAt)

	assert.Equal(t, account.ID, entry.AccountID)
	assert.Equal(t, args.Amount, entry.Amount)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createAndTestRandomAccount(t)
	createAndTestRandomEntry(t, &account)
}

func TestGetEntry(t *testing.T) {
	account := createAndTestRandomAccount(t)
	entry := createAndTestRandomEntry(t, &account)

	actualEntry, err := testQueries.GetEntry(context.Background(), entry.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, actualEntry)

	assert.Equal(t, entry.ID, actualEntry.ID)
	assert.Equal(t, entry.Amount, actualEntry.Amount)
	assert.Equal(t, entry.AccountID, actualEntry.AccountID)
	assert.Equal(t, entry.CreatedAt, actualEntry.CreatedAt)
}

func TestListEntries(t *testing.T) {
	// test only after clearing `entries` table

	account := createAndTestRandomAccount(t)
	entries := make([]Entry, 100)
	for i := range 100 {
		entries[i] = createAndTestRandomEntry(t, &account)
	}

	args := ListEntriesParams{
		AccountID: account.ID,
		Limit:     100,
		Offset:    0,
	}

	actualEntries, err := testQueries.ListEntries(context.Background(), args)
	assert.NoError(t, err)
	assert.Len(t, actualEntries, 100)

	for i, actualEntry := range actualEntries {
		assert.NotEmpty(t, actualEntry)

		entry := entries[i]
		assert.Equal(t, entry.ID, actualEntry.ID)
		assert.Equal(t, entry.Amount, actualEntry.Amount)
		assert.Equal(t, entry.AccountID, actualEntry.AccountID)
		assert.Equal(t, entry.CreatedAt, actualEntry.CreatedAt)
	}
}
