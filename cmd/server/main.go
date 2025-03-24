package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mauzec/simple-bank/api"
	"github.com/mauzec/simple-bank/config"
	db "github.com/mauzec/simple-bank/db/sqlc"
)

// TODO: do restart after changing config
func main() {
	config, err := config.LoadConfig("app", "env", "./config")
	if err != nil {
		log.Fatal("unable to load config:", err)
	}

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal("unable to connect to database:", err)
	}
	defer conn.Close()

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Run(config.ServerAddr)
	if err != nil {
		log.Fatal("catch error when starting server")
	}
}
