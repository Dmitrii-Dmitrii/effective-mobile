package dtos

import "github.com/jackc/pgx/v5/pgtype"

// GetPersonDto @Description Параметры для фильтрации при получении списка людей
type GetPersonDto struct {
	Ids         []pgtype.UUID `json:"ids"`
	Names       []string      `json:"names"`
	Surnames    []string      `json:"surnames"`
	Patronymics []string      `json:"patronymics"`
	LowAge      *uint32       `json:"low_age"`
	HighAge     *uint32       `json:"high_age"`
	Gender      *string       `json:"gender"`
	Countries   []string      `json:"countries"`
	Limit       *uint32       `json:"limit"`
	Offset      *uint32       `json:"offset"`
}
