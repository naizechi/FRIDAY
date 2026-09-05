package request

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	coreerrors "github.com/naizechi/friday/internal/core/errors"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func DecodeAndValidateRequest(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("%w: invalid request body: %w", coreerrors.ErrInvalidArgument, err)
	}

	if err := validate.Struct(dst); err != nil {
		return fmt.Errorf("%w: %w", coreerrors.ErrInvalidArgument, err)
	}

	return nil
}
