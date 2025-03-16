package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbSource = "postgresql://root:secret@localhost:5431/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *pgxpool.Pool // using pool to concurrently run multiple transactions

func TestMain(m *testing.M) {
	ctx := context.Background()
	var err error
	testDB, err = pgxpool.New(ctx, dbSource)
	if err != nil {
		log.Fatal("unable to connect to database", err)
	}
	defer testDB.Close()

	testQueries = New(testDB)

	os.Exit(m.Run())
}
