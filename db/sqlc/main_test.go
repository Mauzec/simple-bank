package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mauzec/simple-bank/config"
)

var testQueries *Queries
var testDB *pgxpool.Pool // using pool to concurrently run multiple transactions

func TestMain(m *testing.M) {
	config, err := config.LoadConfig("app", "env", "../../config")
	if err != nil {
		log.Fatal("unable to load config:", err)
	}

	ctx := context.Background()
	testDB, err = pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal("unable to connect to database", err)
	}
	defer testDB.Close()

	testQueries = New(testDB)

	os.Exit(m.Run())
}
