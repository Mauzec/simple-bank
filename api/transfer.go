package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/token"
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

	authPayload := ctx.MustGet(authPayloadKey).(*token.Payload)

	fromAcc, ok := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if fromAcc.Owner != authPayload.Username {
		err := errors.New("from account does not belong to you")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if !ok {
		return
	}

	_, ok = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !ok {
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

func (server *Server) validAccount(ctx *gin.Context, accountID int64, wantCurrency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return db.Account{}, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return db.Account{}, false
	}

	if wantCurrency == account.Currency {
		return account, true

	}

	ctx.JSON(http.StatusBadRequest, errorResponse(
		fmt.Errorf("account [%d] currency does not match. Want:%s, Got:%s", account.ID, wantCurrency, account.Currency),
	))
	return db.Account{}, false
}
