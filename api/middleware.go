package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mauzec/simple-bank/token"
)

const (
	authHeaderKey  = "auth"
	authPayloadKey = "auth_payload"

	authTypeBearer = "bearer"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(authHeaderKey)
		if len(authHeader) == 0 {
			err := errors.New("auth header is not provided")
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				errorResponse(err),
			)
			return
		}
		fields := strings.Fields(authHeader)
		if len(fields) != 2 {
			err := errors.New("auth header is not accepted")
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				errorResponse(err),
			)
			return
		}

		authType := strings.ToLower(fields[0])
		givenToken := fields[1]
		switch authType {
		case authTypeBearer:
			handleBearer(ctx, tokenMaker, givenToken)

		default:
			err := fmt.Errorf("unsupported auth type %s", authType)
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized,
				errorResponse(err),
			)
			return
		}

		ctx.Next()
	}
}

func handleBearer(ctx *gin.Context, tokenMaker token.Maker, token string) {
	p, err := tokenMaker.VerifyToken(token)
	if err != nil {
		ctx.AbortWithStatusJSON(
			http.StatusUnauthorized,
			errorResponse(err),
		)
		return
	}
	ctx.Set(authPayloadKey, p)
}
