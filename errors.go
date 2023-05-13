package errors

import (
	"errors"
	"fmt"
)

type BusinessError struct {
	HTTPStatusCode int
	Code           string
	Message        string
}

type ResponseError struct {
	BusinessError
	Cause error
}

func (err ResponseError) Error() string {
	return fmt.Sprintf("<ResponseError code: '%s', message: '%s', cause: (%s)>", err.Code, err.Message, err.Cause)
}

//goland:noinspection GoUnusedExportedFunction
func NewResponseError(businessError BusinessError, err error) ResponseError {
	return ResponseError{
		BusinessError: businessError,
		Cause:         err,
	}
}

//goland:noinspection GoUnusedExportedFunction
func ConvertToResponseError(err error, fallbackError BusinessError) ResponseError {
	found := false
	var respError ResponseError
	for {
		if errors.As(err, &respError) {
			found = true
			err = respError.Cause
		} else {
			break
		}
	}

	if found {
		return respError
	}

	return NewResponseError(fallbackError, err)
}
