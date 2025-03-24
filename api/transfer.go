package api

import "github.com/gin-gonic/gin"

type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64   `json:"to_account_id" binding:"required,min=1"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,oneof=EUR USD RUB UAH GBP BYN KZT"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	// TODO: createTransfer
}

func (server *Server) validAccount(ctx *gin.Context) {
	// TODO: valid
}
