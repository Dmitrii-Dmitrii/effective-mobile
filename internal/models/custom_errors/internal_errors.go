package custom_errors

import "fmt"

type InternalError struct {
	Err     error
	Message string
}

func (e *InternalError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

var (
	ErrStartServer    = &InternalError{Message: "failed to start server"}
	ErrShutdownServer = &InternalError{Message: "failed to shutdown server"}
	ErrEnvLoading     = &InternalError{Message: "failed to load .env file loading"}

	ErrBindJsonBody = &InternalError{Message: "failed to bind json body"}

	ErrCreatePool = &InternalError{Message: "failed to create connection pool"}
	ErrScanRow    = &InternalError{Message: "failed to scan row"}

	ErrCreatePerson  = &InternalError{Message: "failed to create person"}
	ErrUpdatePerson  = &InternalError{Message: "failed to update person"}
	ErrGetPersonById = &InternalError{Message: "failed to get person by id"}
	ErrGetPerson     = &InternalError{Message: "failed to get person"}
	ErrDeletePerson  = &InternalError{Message: "failed to delete person"}

	ErrHttpGet = &InternalError{Message: "failed to http get"}

	ErrGetAgeStatusCode    = &InternalError{Message: "failed to get age. status code is not 200"}
	ErrGetAgeReadBody      = &InternalError{Message: "failed to read body while getting age"}
	ErrGetAgeUnmarshalBody = &InternalError{Message: "failed to unmarshal body while getting age"}

	ErrGetGenderStatusCode    = &InternalError{Message: "failed to get gender. status code is not 200"}
	ErrGetGenderReadBody      = &InternalError{Message: "failed to read body while getting gender"}
	ErrGetGenderUnmarshalBody = &InternalError{Message: "failed to unmarshal body while getting gender"}
	ErrGotInvalidGender       = &InternalError{Message: "failed to get gender. invalid gender"}

	ErrGetCountryStatusCode    = &InternalError{Message: "failed to get country. status code is not 200"}
	ErrGetCountryReadBody      = &InternalError{Message: "failed to read body while getting country"}
	ErrGetCountryUnmarshalBody = &InternalError{Message: "failed to unmarshal body while getting country"}
)
