package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/mauzec/simple-bank/db/sqlc"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func (*Server) TODO(ctx *gin.Context) {}

func NewServer(store db.Store) *Server {
	router := gin.Default()
	server := &Server{store: store, router: router}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("currency", validCurrency); err != nil {
			log.Fatal("Something wrong registering currency validate func")
		}
	}

	// single queries

	router.POST("/accounts", server.createAccount)       // createAccount
	router.GET("/accounts/:id", server.getAccount)       // getAccount
	router.GET("/accounts", server.listAccounts)         // listAccounts
	router.DELETE("/accounts/:id", server.deleteAccount) // deleteAccount
	router.PUT("/accounts", server.updateAccount)        // updateAccount

	router.POST("/users", server.createUser) // createUser

	// transactions
	router.POST("/transfers", server.createTransfer)

	return server
}

func (server *Server) Run(addr string) error {
	return server.router.Run(addr)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
