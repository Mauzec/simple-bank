package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

type PSQLStore struct {
	db *pgxpool.Pool
	*Queries
}

func NewStore(db *pgxpool.Pool) Store {
	return &PSQLStore{
		Queries: New(db),
		db:      db,
	}
}

// execTx starts a transaction, creates new Queries with this transaction,
// and executes the function f. If the function returns an error, the transaction is rolled back,
// else the transaction is committed.
func (store *PSQLStore) execTx(ctx context.Context, f func(*Queries) error) error {
	// begin with default transaction options(shortly, just read-commit isolevel)
	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	if err := f(q); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

type txCtxKey string

const txKey txCtxKey = "txName"

func AddBalanceFromTo(
	ctx context.Context,
	q *Queries,
	fromAccID int64,
	toAccID int64,
	amount1 float64,
	amount2 float64,
) (acc1 Account, acc2 Account, err error) {
	acc1, err = q.AddBalanceAccount(ctx, AddBalanceAccountParams{
		ID:     fromAccID,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	acc2, err = q.AddBalanceAccount(ctx, AddBalanceAccountParams{
		ID:     toAccID,
		Amount: amount2,
	})
	if err != nil {
		return
	}

	return
}

// TransferTx create a transfer record to do a money transfer between two accounts.
// It create a new row in the transfer table, add two records in the entries table, and
// update the accounts' balance within a single database transaction.
func (store *PSQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// update accounts' balance

		// to avoid deadlocks
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = AddBalanceFromTo(ctx, q,
				arg.FromAccountID,
				arg.ToAccountID,
				-arg.Amount,
				arg.Amount,
			)
		} else {
			result.ToAccount, result.FromAccount, err = AddBalanceFromTo(ctx, q,
				arg.ToAccountID,
				arg.FromAccountID,
				arg.Amount,
				-arg.Amount,
			)
		}

		return err
	})

	return result, err
}

type TransferTxParams struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}
