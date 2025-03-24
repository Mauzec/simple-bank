package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/mauzec/simple-bank/db/sqlc"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

func (*Server) TODO(ctx *gin.Context) {}

func NewServer(store *db.Store) *Server {
	router := gin.Default()
	server := &Server{store: store, router: router}

	router.POST("/accounts", server.createAccount)       // createAccount
	router.GET("/accounts/:id", server.getAccount)       // getAccount
	router.GET("/accounts", server.listAccounts)         // listAccounts
	router.DELETE("/accounts/:id", server.deleteAccount) // deleteAccount
	router.PUT("/accounts", server.updateAccount)        // updateAccount

	return server
}

func (server *Server) Run(addr string) error {
	return server.router.Run(addr)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
