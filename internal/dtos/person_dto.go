package dtos

import "github.com/jackc/pgx/v5/pgtype"

type PersonDto struct {
	Id         *pgtype.UUID `db:"id"`
	Name       *string      `db:"name,omitempty"`
	Surname    *string      `db:"surname,omitempty"`
	Patronymic *string      `db:"patronymic,omitempty"`
	Age        *int         `db:"age,omitempty"`
	Gender     *string      `db:"gender,omitempty"`
	Country    *string      `db:"country,omitempty"`
}
