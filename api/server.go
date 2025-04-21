package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/token"
)

type Server struct {
	store  db.Store
	router *gin.Engine

	tokenMaker  token.Maker
	tokenParams TokenParams
}

func (*Server) TODO(ctx *gin.Context) {}

// TODO: delete it or add something else in this structure later
type TokenParams struct {
	AccessTokenDuration time.Duration
}

func NewServer(store db.Store, tokenMaker token.Maker, tokenParams TokenParams) (*Server, error) {
	server := &Server{store: store, tokenMaker: tokenMaker, tokenParams: tokenParams}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("currency", validCurrency); err != nil {
			return nil, fmt.Errorf("something wrong registering currency validate func")
		}
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	// single queries
	router.POST("/users", server.createUser)      // createUser
	router.POST("/users/login", server.loginUser) // loginUser

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/accounts", server.createAccount) // createAccount
	authRoutes.GET("/accounts/:id", server.getAccount) // getAccount
	authRoutes.GET("/accounts", server.listAccounts)   // listAccounts

	// dangerous single quries
	// ONLY FOR ADMINS
	// TODO: craete roles
	authRoutes.PUT("/accounts", server.updateAccount)        // updateAccount
	authRoutes.DELETE("/accounts/:id", server.deleteAccount) // deleteAccount

	// transactions
	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router
}

func (server *Server) Run(addr string) error {
	return server.router.Run(addr)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
