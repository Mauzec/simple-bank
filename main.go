package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mauzec/simple-bank/api"
	db "github.com/mauzec/simple-bank/db/sqlc"
)

const (
	dbSource   = "postgresql://root:secret@localhost:5431/simple_bank?sslmode=disable"
	serverAddr = "127.0.0.1:8080"
)

func main() {
	conn, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("unable to connect to database", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Run(serverAddr)
	if err != nil {
		log.Fatal("catch error when starting server")
	}
}
