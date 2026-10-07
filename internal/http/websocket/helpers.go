package websocket

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func validateRequest(v *validator.Validate, request Request) error {
	if err := v.Struct(request); err != nil {
		return err
	}

	if !validRequestID(request.RID) {
		return fmt.Errorf("request id must be a string or number")
	}

	return nil
}

func validRequestID(value any) bool {
	switch value.(type) {
	case string, float64:
		return true
	default:
		return false
	}
}
