package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mauzec/simple-bank/api"
	"github.com/mauzec/simple-bank/config"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/token"
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

	var tokenMaker token.Maker
	switch config.TokenType {
	case "JWTS":
		tokenMaker, err = token.NewJWTSMaker(config.TokenSymmetricKey)
	case "PasetoS":
		tokenMaker, err = token.NewPasetoSMaker(config.TokenSymmetricKey)
	default:
		log.Fatal("given unsupported token type")
	}
	if err != nil {
		log.Fatal("something go wrong when creating token maker")
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(store, tokenMaker, api.TokenParams{
		AccessTokenDuration: config.AccessTokenDuration,
	})
	if err != nil {
		log.Fatal("server creating err:", err)
	}

	err = server.Run(config.ServerAddr)
	if err != nil {
		log.Fatal("catch error when starting server")
	}
}
