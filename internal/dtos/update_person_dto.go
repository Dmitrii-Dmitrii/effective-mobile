package dtos

import "github.com/jackc/pgx/v5/pgtype"

type UpdatePersonDto struct {
	Id         pgtype.UUID `json:"id" db:"id"`
	Name       *string     `json:"name,omitempty" db:"name,omitempty"`
	Surname    *string     `json:"surname,omitempty" db:"surname,omitempty"`
	Patronymic *string     `json:"patronymic,omitempty" db:"patronymic,omitempty"`
	Age        *int        `json:"age,omitempty" db:"age,omitempty"`
	Gender     *string     `json:"gender,omitempty" db:"gender,omitempty"`
	Country    *string     `json:"country,omitempty" db:"country,omitempty"`
}
