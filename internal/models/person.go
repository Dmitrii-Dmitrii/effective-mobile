package models

import "github.com/jackc/pgx/v5/pgtype"

type Person struct {
	id         pgtype.UUID
	name       string
	surname    string
	patronymic string
	age        uint32
	gender     GenderType
	country    string
}

type GenderType string

const (
	male   GenderType = "male"
	female GenderType = "female"
)
