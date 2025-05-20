package services

import (
	"context"
	"effective-mobile/internal/drivers"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"io"
	"net/http"
	"os"
)

type PersonService struct {
	personDriver drivers.PersonDriver
}

func NewPersonService(personDriver drivers.PersonDriver) *PersonService {
	return &PersonService{personDriver: personDriver}
}

func (s *PersonService) CreatePerson(ctx context.Context, personDto dtos.CreatePersonDto) (*dtos.PersonDto, error) {
	personId := generateUuid()

	ageChan := make(chan uint32, 1)
	ageErrChan := make(chan error, 1)

	genderChan := make(chan models.GenderType, 1)
	genderErrChan := make(chan error, 1)

	countryChan := make(chan string, 1)
	countryErrChan := make(chan error, 1)

	go func() {
		age, err := getAge(personDto.Name)
		ageChan <- age
		ageErrChan <- err
	}()

	go func() {
		gender, err := getGender(personDto.Name)
		genderChan <- gender
		genderErrChan <- err
	}()

	go func() {
		country, err := getCountry(personDto.Name)
		countryChan <- country
		countryErrChan <- err
	}()

	age := <-ageChan
	ageErr := <-ageErrChan
	if ageErr != nil {
		return nil, ageErr
	}

	gender := <-genderChan
	genderErr := <-genderErrChan
	if genderErr != nil {
		return nil, genderErr
	}

	country := <-countryChan
	countryErr := <-countryErrChan
	if countryErr != nil {
		return nil, countryErr
	}

	person := &models.Person{Id: personId, Name: personDto.Name, Surname: personDto.Surname, Age: age, Country: country, Gender: gender}
	if personDto.Patronymic != nil {
		person.Patronymic = *personDto.Patronymic
	}

	if err := s.personDriver.CreatePerson(ctx, person); err != nil {
		return nil, err
	}

	createdPersonDto := mapPersonToDto(person)

	return createdPersonDto, nil
}

func (s *PersonService) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*dtos.PersonDto, error) {
	_, err := s.GetPersonById(ctx, personDto.Id)
	if err != nil {
		return nil, err
	}

	updatedPerson, err := s.personDriver.UpdatePerson(ctx, personDto)
	if err != nil {
		return nil, err
	}

	genderDto := string(updatedPerson.Gender)

	updatedPersonDto := &dtos.PersonDto{
		Id:         updatedPerson.Id,
		Name:       &updatedPerson.Name,
		Surname:    &updatedPerson.Surname,
		Patronymic: &updatedPerson.Patronymic,
		Age:        &updatedPerson.Age,
		Gender:     &genderDto,
		Country:    &updatedPerson.Country,
	}
	return updatedPersonDto, nil
}

func (s *PersonService) DeletePerson(ctx context.Context, personId pgtype.UUID) error {
	err := s.personDriver.DeletePerson(ctx, personId)
	return err
}

func (s *PersonService) GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]dtos.PersonDto, error) {
	if err := validateGetPersonDto(getPersonDto); err != nil {
		return nil, err
	}

	persons, err := s.personDriver.GetPersons(ctx, getPersonDto)
	if err != nil {
		return nil, err
	}

	personDtos := make([]dtos.PersonDto, len(persons))
	for i, person := range persons {
		personDtos[i] = *mapPersonToDto(&person)
	}

	return personDtos, nil
}

func (s *PersonService) GetPersonById(ctx context.Context, personId pgtype.UUID) (*dtos.PersonDto, error) {
	person, err := s.personDriver.GetPersonById(ctx, personId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("person not found")
	}
	if err != nil {
		return nil, err
	}

	personDto := mapPersonToDto(person)

	return personDto, nil
}

func getAge(name string) (uint32, error) {
	ageUrl := os.Getenv("AGE_URL") + name

	resp, err := http.Get(ageUrl)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var ageDto dtos.AgeDto
	if err = json.Unmarshal(body, &ageDto); err != nil {
		return 0, err
	}

	return ageDto.Age, nil
}

func getGender(name string) (models.GenderType, error) {
	ageUrl := os.Getenv("GENDER_URL") + name

	resp, err := http.Get(ageUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var genderDto dtos.GenderDto
	if err = json.Unmarshal(body, &genderDto); err != nil {
		return "", err
	}

	gender := models.GenderType(genderDto.Gender)
	if gender != models.Male && gender != models.Female {
		return "", errors.New("invalid gender")
	}

	return gender, nil
}

func getCountry(name string) (string, error) {
	ageUrl := os.Getenv("COUNTRY_URL") + name

	resp, err := http.Get(ageUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var countryDto dtos.CountryDto
	if err = json.Unmarshal(body, &countryDto); err != nil {
		return "", err
	}

	maxProb := countryDto.Countries[0]
	for _, country := range countryDto.Countries {
		if country.Probability > maxProb.Probability {
			maxProb = country
		}
	}

	return maxProb.Id, nil
}

func generateUuid() pgtype.UUID {
	newUuid := uuid.New()

	pgUuid := pgtype.UUID{
		Bytes: newUuid,
		Valid: true,
	}

	return pgUuid
}

func mapPersonToDto(person *models.Person) *dtos.PersonDto {
	genderDto := string(person.Gender)

	personDto := &dtos.PersonDto{
		Id:         person.Id,
		Name:       &person.Name,
		Surname:    &person.Surname,
		Patronymic: &person.Patronymic,
		Age:        &person.Age,
		Gender:     &genderDto,
		Country:    &person.Country,
	}

	return personDto
}

func validateGetPersonDto(getPersonDto dtos.GetPersonDto) error {
	if len(getPersonDto.Ids) > 0 {
		for _, id := range getPersonDto.Ids {
			if !id.Valid {
				return errors.New("invalid person id")
			}
		}
	}

	if getPersonDto.Limit != nil && *getPersonDto.Limit < 0 {
		return errors.New("limit cannot be negative")
	}

	if getPersonDto.Offset != nil && *getPersonDto.Offset < 0 {
		return errors.New("offset cannot be negative")
	}

	if getPersonDto.LowAge != nil && *getPersonDto.LowAge < 0 {
		return errors.New("low age cannot be negative")
	}

	if getPersonDto.HighAge != nil && *getPersonDto.HighAge < 0 {
		return errors.New("high age cannot be negative")
	}

	if getPersonDto.Gender != nil && *getPersonDto.Gender != "male" && *getPersonDto.Gender != "female" {
		return errors.New("gender must be either 'male' or 'female'")
	}

	return nil
}
