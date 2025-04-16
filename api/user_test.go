package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	mockdb "github.com/mauzec/simple-bank/db/mock"
	db "github.com/mauzec/simple-bank/db/sqlc"
	"github.com/mauzec/simple-bank/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(16)
	hashedPassword, err := util.HashPassword(password)
	assert.NoError(t, err)
	user := db.User{
		Username:       util.RandomOwner(),
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
	}

	return user, password
}

type EqUserMatcher struct {
	x        db.CreateUserParams
	password string
}

func (e *EqUserMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(arg.HashedPassword, e.password)
	if err != nil {
		return false
	}

	e.x.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.x, arg)
}

func (e *EqUserMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.x, e.password)
}

func EqUser(arg db.CreateUserParams, password string) *EqUserMatcher {
	return &EqUserMatcher{arg, password}
}

func assertBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	assert.NoError(t, err)
	var leftUser db.User
	err = json.Unmarshal(data, &leftUser)
	assert.NoError(t, err)
	leftUser.HashedPassword = user.HashedPassword
	assert.Equal(t, user, leftUser)
}

func TestCreateUserAPI(t *testing.T) {
	rndUser, rndPassword := randomUser(t)

	testCases := []struct {
		testName      string
		bodyReq       gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			"OK",
			gin.H{
				"username":  rndUser.Username,
				"full_name": rndUser.FullName,
				"password":  rndPassword,
				"email":     rndUser.Email,
			},
			func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: rndUser.Username,
					FullName: rndUser.FullName,
					Email:    rndUser.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqUser(arg, rndPassword)).
					Times(1).
					Return(rndUser, nil)
			},
			func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				assertBodyMatchUser(t, recorder.Body, rndUser)
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

			data, err := json.Marshal(tc.bodyReq)
			assert.NoError(t, err)

			url := "/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
			assert.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
