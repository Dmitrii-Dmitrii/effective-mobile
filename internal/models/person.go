package models

import "github.com/jackc/pgx/v5/pgtype"

type Person struct {
	Id         pgtype.UUID
	Name       string
	Surname    string
	Patronymic string
	Age        uint32
	Gender     GenderType
	Country    string
}

type GenderType string

const (
	Male   GenderType = "male"
	Female GenderType = "female"
)
