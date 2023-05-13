package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"strings"
	"testing"
)

var (
	ErrorSuccess = BusinessError{HTTPStatusCode: http.StatusOK, Code: "", Message: ""}

	ErrorInvalidData = BusinessError{HTTPStatusCode: http.StatusBadRequest, Code: "USER-00000010", Message: "invalid request body"}

	ErrorCreateUser      = BusinessError{Code: "USER-00010000", Message: "failed to create user"}
	ErrorMissingUsername = BusinessError{HTTPStatusCode: http.StatusBadRequest, Code: "USER-00010020", Message: "missing username"}
	ErrorUserExist       = BusinessError{Code: "USER-00010030", Message: "user already exists"}

	ErrorUnknown = BusinessError{Code: "USER-10000010", Message: "unknown error"}
)

func TestResponseError(t *testing.T) {
	suite.Run(t, new(TestSuiteResponseError))
}

type TestSuiteResponseError struct {
	suite.Suite
}

func (suite *TestSuiteResponseError) TestResponseError() {
	args := []struct {
		name             string
		data             string
		expectedHTTPCode int
		expectedError    BusinessError
	}{
		{
			name:             "success",
			data:             `{"Username": "zhang san", "Email": "zhang.san@company.com"}`,
			expectedHTTPCode: http.StatusOK,
			expectedError:    ErrorSuccess,
		},
		{
			name:             "invalid body",
			data:             `{"Username": "zhang san"`,
			expectedHTTPCode: http.StatusBadRequest,
			expectedError:    ErrorInvalidData,
		},
		{
			name:             "missing username",
			data:             `{"Username": "", "Email": "user@company.com"}`,
			expectedHTTPCode: http.StatusBadRequest,
			expectedError:    ErrorMissingUsername,
		},
		{
			name:             "user exist",
			data:             `{"Username": "existed_user", "Email": "user@company.com"}`,
			expectedHTTPCode: http.StatusInternalServerError,
			expectedError:    ErrorUserExist,
		},
		{
			name:             "create user failed",
			data:             `{"Username": "failed_user", "Email": "user@company.com"}`,
			expectedHTTPCode: http.StatusInternalServerError,
			expectedError:    ErrorCreateUser,
		},
	}
	for _, arg := range args {
		httpCode, _, errorCode, errorMessage := handleRequest(arg.data)
		assert.Equal(suite.T(), arg.expectedHTTPCode, httpCode)
		assert.Equal(suite.T(), arg.expectedError.Code, errorCode)
		assert.Equal(suite.T(), arg.expectedError.Message, errorMessage)
	}
}

func handleRequest(data string) (int, interface{}, string, string) {
	resp, err := controllerCreateUser(data)
	if err != nil {
		responseError := ConvertToResponseError(err, ErrorUnknown)

		httpStatusCode := responseError.HTTPStatusCode
		if httpStatusCode == 0 {
			httpStatusCode = http.StatusInternalServerError
		}

		fmt.Printf("request failed: (%v)\n", err)

		return httpStatusCode, nil, responseError.Code, responseError.Message
	}

	return http.StatusOK, resp, "", ""
}

func controllerCreateUser(data string) (interface{}, error) {
	var req CreateUserRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return nil, NewResponseError(ErrorInvalidData, fmt.Errorf("unmarshal request data:(%w)", err))
	}

	resp, err := serviceCreateUser(req)
	if err != nil {
		return nil, NewResponseError(ErrorCreateUser, fmt.Errorf("service create user: (%w)", err))
	}

	return resp, nil
}

func serviceCreateUser(req CreateUserRequest) (CreateUserResponse, error) {
	if len(req.Username) == 0 {
		return CreateUserResponse{}, NewResponseError(ErrorMissingUsername, errors.New("username was empty"))
	}

	existedUser, err := repositoryGetUserByName(req.Username)
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("repository get user by name: (%w)", err)
	}

	if len(existedUser.Username) > 0 {
		return CreateUserResponse{}, NewResponseError(ErrorUserExist, errors.New("user already exist"))
	}

	user, err := repositoryCreateUser(User{Username: req.Username, Email: req.Email})
	if err != nil {
		return CreateUserResponse{}, fmt.Errorf("repository get user by name: (%w)", err)
	}

	return CreateUserResponse{ID: user.ID, Username: user.Username, Email: user.Email}, nil
}

func repositoryGetUserByName(username string) (User, error) {
	// mock db search
	if strings.HasPrefix(username, "exist") {
		return User{ID: 1, Username: username, Email: "email"}, nil
	}

	return User{}, nil
}

func repositoryCreateUser(user User) (User, error) {
	if strings.HasPrefix(user.Username, "failed") {
		return User{}, errors.New("database error")
	}

	return User{ID: 100, Username: user.Username, Email: user.Email}, nil
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type CreateUserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}
