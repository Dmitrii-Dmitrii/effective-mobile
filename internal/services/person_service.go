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

func (s *PersonService) CreatePerson(ctx context.Context, personDto dtos.CreatePersonDto) (*models.Person, error) {
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

	return person, nil
}

func (s *PersonService) UpdatePerson(ctx context.Context, personDto dtos.UpdatePersonDto) (*models.Person, error) {
	person, err := s.GetPersonById(ctx, personDto.Id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("person not found")
	}
	if err != nil {
		return nil, err
	}

}

func (s *PersonService) GetPersonById(ctx context.Context, personId pgtype.UUID) (*models.Person, error) {
	person, err := s.personDriver.GetPersonById(ctx, personId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("person not found")
	}
	if err != nil {
		return nil, err
	}

	return person, nil
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
