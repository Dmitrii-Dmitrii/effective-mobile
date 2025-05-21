package dtos

// CreatePersonDto @Description Данные для создания новой записи о человеке
type CreatePersonDto struct {
	Name       string  `json:"name"`
	Surname    string  `json:"surname"`
	Patronymic *string `json:"patronymic,omitempty"`
}
