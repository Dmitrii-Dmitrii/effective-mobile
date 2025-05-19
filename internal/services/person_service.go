package services

import (
	"effective-mobile/internal/drivers"
)

type PersonService struct {
	personDriver drivers.PersonDriver
}

func NewPersonService(personDriver drivers.PersonDriver) *PersonService {
	return &PersonService{personDriver: personDriver}
}

//func (service *PersonService) CreatePerson(person dtos.PersonDto) (models.Person, error) {}
