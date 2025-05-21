package drivers

import (
	"context"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

type PersonDriverInterface interface {
	CreatePerson(ctx context.Context, person *models.Person) error
	UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*models.Person, error)
	DeletePerson(ctx context.Context, personId pgtype.UUID) error
	GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]models.Person, error)
	GetPersonById(ctx context.Context, id pgtype.UUID) (*models.Person, error)
}
