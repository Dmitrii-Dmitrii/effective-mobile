package custom_errors

import "fmt"

type UserError struct {
	Err     error
	Message string
}

func (e *UserError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

var (
	ErrPersonNotFound   = &UserError{Message: "person not found"}
	ErrNoFieldsToUpdate = &InternalError{Message: "no fields to update"}

	ErrInvalidUuid   = &InternalError{Message: "invalid UUID"}
	ErrLimitValue    = &UserError{Message: "limit cannot be negative"}
	ErrOffsetValue   = &UserError{Message: "offset cannot be negative"}
	ErrLowAgeValue   = &UserError{Message: "low age cannot be negative"}
	ErrHighAgeValue  = &UserError{Message: "high age cannot be negative"}
	ErrInvalidGender = &UserError{Message: "gender must be either 'male' or 'female'"}
)
