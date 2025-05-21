package dtos

import "github.com/jackc/pgx/v5/pgtype"

// PersonDto @Description Полная информация о человеке
type PersonDto struct {
	Id         pgtype.UUID `json:"id"`
	Name       *string     `json:"name,omitempty"`
	Surname    *string     `json:"surname,omitempty"`
	Patronymic *string     `json:"patronymic,omitempty"`
	Age        *uint32     `json:"age,omitempty"`
	Gender     *string     `json:"gender,omitempty"`
	Country    *string     `json:"country,omitempty"`
}
