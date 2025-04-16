package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/token"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func newTestServer(t *testing.T, store db.Store) *Server {
	tokenMaker, err := token.NewPasetoSMaker("12345678901234567890123456789012")
	assert.NoError(t, err)

	tokenParams := TokenParams{time.Minute * 15}
	server, err := NewServer(store, tokenMaker, tokenParams)
	assert.NoError(t, err)

	return server
}
