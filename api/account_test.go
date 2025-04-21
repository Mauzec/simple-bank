package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	mockdb "github.com/mauzec/simple-bank/db/mock"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/token"
	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 10000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}
}
func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()
	account2 := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					account.Owner,
					time.Minute*15)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					account.Owner,
					time.Minute*15,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					account.Owner,
					time.Minute*15,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					account.Owner,
					time.Minute*15,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name:      "AccountNotBelongToUser",
			accountID: account2.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					account.Owner,
					time.Minute*15,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
					Times(1).
					Return(account2, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuth",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccountAPI(t *testing.T) {
	rndAcc := db.Account{
		Owner:    util.RandomOwner(),
		Balance:  0,
		Currency: util.RandomCurrency(),
	}

	createAccountParams := db.CreateAccountParams{
		Owner:    rndAcc.Owner,
		Balance:  rndAcc.Balance,
		Currency: rndAcc.Currency,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "InvalidRequestNoCurrency",
			body: gin.H{},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					rndAcc.Owner,
					time.Minute,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"currency": "HEY",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					rndAcc.Owner,
					time.Minute,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"currency": rndAcc.Currency,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					rndAcc.Owner,
					time.Minute,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(createAccountParams)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},

		{
			name: "OK",
			body: gin.H{
				"currency": rndAcc.Currency,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker,
					authTypeBearer,
					rndAcc.Owner,
					time.Minute,
				)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(createAccountParams)).
					Times(1).
					Return(rndAcc, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertBodyMatchAccount(t, recorder.Body, rndAcc)
			},
		},
		{
			name: "NoAuth",
			body: map[string]any{
				"currency": rndAcc.Currency,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(createAccountParams)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			url := "/accounts"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteAccountAPI(t *testing.T) {
	acc1 := randomAccount()
	acc2 := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "InvalidRequestBadID",
			accountID: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, acc1.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: acc1.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, acc1.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(acc1, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "AccountNotBelongToUser",
			accountID: acc2.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, acc1.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(1).
					Return(acc2, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(acc2.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "OK",
			accountID: acc1.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, acc1.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(acc1, nil)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: acc1.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, acc1.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc1.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "NoAuth",
			accountID: acc1.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestUpdateAccountAPI(t *testing.T) {
	rndAcc := randomAccount()
	updAcc := rndAcc
	updAcc.Balance = 2077

	updAccParams := db.UpdateAccountParams{
		ID:      rndAcc.ID,
		Balance: 2077,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "InvalidRequestNoID",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			body: gin.H{
				"balance": updAccParams.Balance,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "InvalidRequestNoBalance",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			body: gin.H{
				"id": rndAcc.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "InvalidRequestZeroBody",
			body: gin.H{},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "OK",
			body: gin.H{
				"id":      rndAcc.ID,
				"balance": updAccParams.Balance,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(rndAcc.ID)).
					Times(1).
					Return(rndAcc, nil)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(updAccParams)).
					Times(1).
					Return(updAcc, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertBodyMatchAccount(t, recorder.Body, updAcc)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"id":      rndAcc.ID,
				"balance": updAccParams.Balance,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(rndAcc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(updAccParams)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"id":      rndAcc.ID,
				"balance": updAccParams.Balance,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(rndAcc.ID)).
					Times(1).
					Return(rndAcc, nil)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(updAccParams)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "NoAuth",
			body: gin.H{
				"id":      rndAcc.ID,
				"balance": updAccParams.Balance,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
		{
			name: "AccountNotBelongToUser",
			body: gin.H{
				"id":      int64(1),
				"balance": updAccParams.Balance,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, rndAcc.Owner, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(int64(1))).
					Times(1).
					Return(db.Account{ID: 1, Owner: "hello"}, nil)
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assertBodyNoAccount(t, recorder.Body)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			assert.NoError(t, err)

			url := "/accounts"
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	accounts := make([]db.Account, 5)
	username := "hello"
	for i := range accounts {
		accounts[i] = randomAccount()
		accounts[i].Owner = username
	}
	n := int32(5)

	testCases := []struct {
		name          string
		pageNumber    int32
		pageSize      int32
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "InvalidPageNumber",
			pageNumber: 0,
			pageSize:   30,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:       "InvalidPageSize",
			pageSize:   100,
			pageNumber: 1,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:       "InvalidNoParameters",
			pageSize:   -1,
			pageNumber: -1,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},

		{
			name:       "OK",
			pageSize:   n,
			pageNumber: 25,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Owner:  username,
						Limit:  n,
						Offset: n * (25 - 1),
					})).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},

		{
			name:       "InternalError",
			pageSize:   n,
			pageNumber: 32,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, authTypeBearer, username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{
						Owner:  username,
						Limit:  n,
						Offset: n * (32 - 1),
					})).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, recorder.Code)
				assertBodyNoAccounts(t, recorder.Body)
			},
		},
		{
			name:       "NoAuth",
			pageSize:   n,
			pageNumber: 32,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusUnauthorized, recorder.Code)
				assertBodyNoAccounts(t, recorder.Body)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			var url string
			if tc.pageNumber == -1 && tc.pageSize == -1 {
				url = "/accounts"
			} else {
				url = fmt.Sprintf("/accounts?page_number=%d&page_size=%d", tc.pageNumber, tc.pageSize)
			}
			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func assertBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)
	var leftAcc db.Account
	err = json.Unmarshal(data, &leftAcc)
	assert.NoError(t, err)
	assert.Equal(t, account, leftAcc)
}

func assertBodyMatchAccounts(t *testing.T, body *bytes.Buffer, account []db.Account) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)
	var leftAccs []db.Account
	err = json.Unmarshal(data, &leftAccs)
	assert.NoError(t, err)
	assert.Equal(t, account, leftAccs)
}

func assertBodyNoAccount(t *testing.T, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var acc db.Account
	err = json.Unmarshal(data, &acc)
	assert.NoError(t, err)
	assert.Zero(t, acc)
}

func assertBodyNoAccounts(t *testing.T, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)

	var accounts []db.Account
	_ = json.Unmarshal(data, &accounts)
	// TODO: todo if response will be a map type
	// now here unmarshal error, because tested for response like "error": "... connection refused ..."
	// assert.NoError(t, err)
	assert.Zero(t, accounts)
}
