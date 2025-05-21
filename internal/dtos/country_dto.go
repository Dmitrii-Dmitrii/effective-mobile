package dtos

type CountryDto struct {
	Countries []CountryIdDto `json:"country"`
}

type CountryIdDto struct {
	Id          string  `json:"country_id"`
	Probability float64 `json:"probability"`
}
