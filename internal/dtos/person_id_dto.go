package dtos

import "github.com/google/uuid"

type PersonIdDto struct {
	PersonId uuid.UUID `json:"id"`
}
