package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

const (
	dbSource = "postgresql://root:secret@localhost:5431/simple_bank?sslmode=disable"
)

var testQueries *Queries

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
