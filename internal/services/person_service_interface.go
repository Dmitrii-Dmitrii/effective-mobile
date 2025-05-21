package services

import (
	"context"
	"effective-mobile/internal/dtos"
	"github.com/jackc/pgx/v5/pgtype"
)

type PersonServiceInterface interface {
	CreatePerson(ctx context.Context, personDto dtos.CreatePersonDto) (*dtos.PersonDto, error)
	UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*dtos.PersonDto, error)
	DeletePerson(ctx context.Context, personId pgtype.UUID) error
	GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]dtos.PersonDto, error)
	GetPersonById(ctx context.Context, personId pgtype.UUID) (*dtos.PersonDto, error)
}
