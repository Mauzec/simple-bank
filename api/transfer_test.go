package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	mockdb "github.com/mauzec/simple-bank/db/mock"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTransferAPI(t *testing.T) {
	acc1 := randomAccount()
	acc2 := randomAccount()
	acc3 := randomAccount()

	acc1.Currency, acc2.Currency, acc3.Currency = util.USD, util.USD, util.KZT

	testCases := []struct {
		testName      string
		bodyResponse  gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			testName: "OK",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          100,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(acc1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).
					Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        100,
					})).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			testName: "NoToAccIDField",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"amount":          100,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(0).
					Return(acc1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0).
					Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        100,
					})).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "NoFromAccIDField",
			bodyResponse: gin.H{
				"to_account_id": acc2.ID,
				"amount":        100,
				"currency":      util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(0).
					Return(acc1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0).
					Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        100,
					})).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "NoAmountField",
			bodyResponse: gin.H{
				"from_to_account_id": acc1.ID,
				"to_account_id":      acc2.ID,
				"currency":           util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(0).
					Return(acc1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0).
					Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        100,
					})).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "FromAccCurrenciesMismath",
			bodyResponse: gin.H{
				"from_account_id": acc3.ID,
				"to_account_id":   acc1.ID,
				"amount":          100,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc3.ID)).
					Times(1).
					Return(acc3, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), 100).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "ToAccCurrenciesMismath",
			bodyResponse: gin.H{
				"from_account_id": acc3.ID,
				"to_account_id":   acc1.ID,
				"amount":          100,
				"currency":        util.KZT,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc3.ID)).
					Times(1).
					Return(acc3, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(acc1, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), 100).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "BadCurrency",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          100,
				"currency":        "HEY",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "NegativeAmount",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          -0.01,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			testName: "AccountNotFound",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          100,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			testName: "GetAccountInternalError",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          100,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(acc1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			testName: "TransferTxInternalError",
			bodyResponse: gin.H{
				"from_account_id": acc1.ID,
				"to_account_id":   acc2.ID,
				"amount":          100,
				"currency":        util.USD,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(acc1, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).
					Return(acc2, nil)

				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: acc1.ID,
						ToAccountID:   acc2.ID,
						Amount:        100,
					})).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.bodyResponse)
			assert.NoError(t, err)

			url := "/transfers"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
