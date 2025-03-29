package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mauzec/simple-bank/db/sqlc"
)

type TransferRequest struct {
	FromAccountID int64   `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64   `json:"to_account_id" binding:"required,min=1"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.matchAccount(ctx, req.FromAccountID, req.Currency) ||
		!server.matchAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	transferResult, err := server.store.TransferTx(ctx, db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, transferResult)
}

func (server *Server) matchAccount(ctx *gin.Context, accountID int64, wantCurrency string) bool {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if wantCurrency == account.Currency {
		return true
	}

	ctx.JSON(http.StatusBadRequest, errorResponse(
		fmt.Errorf("account [%d] currency does not match. Want:%s, Got:%s", account.ID, wantCurrency, account.Currency),
	))
	return false
}
