package services

import (
	"context"
	"effective-mobile/internal/drivers"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"effective-mobile/internal/models/custom_errors"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

type PersonService struct {
	personDriver drivers.PersonDriverInterface
}

func NewPersonService(personDriver drivers.PersonDriverInterface) *PersonService {
	log.Debug().Msg("Initializing PersonService")
	return &PersonService{personDriver: personDriver}
}

func (s *PersonService) CreatePerson(ctx context.Context, personDto dtos.CreatePersonDto) (*dtos.PersonDto, error) {
	log.Info().
		Str("name", personDto.Name).
		Str("surname", personDto.Surname).
		Bool("has_patronymic", personDto.Patronymic != nil).
		Msg("Creating new person")

	personId := generateUuid()
	log.Debug().Str("generated_uuid", personId.String()).Msg("Generated UUID for new person")

	ageChan := make(chan uint32, 1)
	ageErrChan := make(chan error, 1)

	genderChan := make(chan models.GenderType, 1)
	genderErrChan := make(chan error, 1)

	countryChan := make(chan string, 1)
	countryErrChan := make(chan error, 1)

	log.Debug().Msg("Starting goroutines to fetch person attributes")

	go func() {
		log.Debug().Str("name", personDto.Name).Msg("Fetching age")
		age, err := getAge(personDto.Name)
		ageChan <- age
		ageErrChan <- err
		if err == nil {
			log.Debug().Str("name", personDto.Name).Uint32("age", age).Msg("Age fetched successfully")
		}
	}()

	go func() {
		log.Debug().Str("name", personDto.Name).Msg("Fetching gender")
		gender, err := getGender(personDto.Name)
		genderChan <- gender
		genderErrChan <- err
		if err == nil {
			log.Debug().Str("name", personDto.Name).Str("gender", string(gender)).Msg("Gender fetched successfully")
		}
	}()

	go func() {
		log.Debug().Str("name", personDto.Name).Msg("Fetching country")
		country, err := getCountry(personDto.Name)
		countryChan <- country
		countryErrChan <- err
		if err == nil {
			log.Debug().Str("name", personDto.Name).Str("country", country).Msg("Country fetched successfully")
		}
	}()

	log.Debug().Msg("Waiting for age result")
	age := <-ageChan
	ageErr := <-ageErrChan
	if ageErr != nil {
		log.Error().
			Err(ageErr).
			Str("name", personDto.Name).
			Msg("Failed to get age")
		return nil, ageErr
	}

	log.Debug().Msg("Waiting for gender result")
	gender := <-genderChan
	genderErr := <-genderErrChan
	if genderErr != nil {
		log.Error().
			Err(genderErr).
			Str("name", personDto.Name).
			Msg("Failed to get gender")
		return nil, genderErr
	}

	log.Debug().Msg("Waiting for country result")
	country := <-countryChan
	countryErr := <-countryErrChan
	if countryErr != nil {
		log.Error().
			Err(countryErr).
			Str("name", personDto.Name).
			Msg("Failed to get country")
		return nil, countryErr
	}

	log.Debug().
		Str("person_id", personId.String()).
		Str("name", personDto.Name).
		Str("surname", personDto.Surname).
		Uint32("age", age).
		Str("country", country).
		Str("gender", string(gender)).
		Msg("Prepared person data")

	person := &models.Person{Id: personId, Name: personDto.Name, Surname: personDto.Surname, Age: age, Country: country, Gender: gender}
	if personDto.Patronymic != nil {
		person.Patronymic = *personDto.Patronymic
		log.Debug().
			Str("person_id", personId.String()).
			Str("patronymic", *personDto.Patronymic).
			Msg("Added patronymic to person")
	}

	log.Debug().Str("person_id", personId.String()).Msg("Saving person to database")
	if err := s.personDriver.CreatePerson(ctx, person); err != nil {
		log.Error().
			Err(err).
			Str("person_id", personId.String()).
			Msg("Failed to save person to database")
		return nil, err
	}

	log.Info().
		Str("person_id", personId.String()).
		Str("name", person.Name).
		Str("surname", person.Surname).
		Uint32("age", person.Age).
		Str("country", person.Country).
		Str("gender", string(person.Gender)).
		Msg("Person created successfully")

	createdPersonDto := mapPersonToDto(person)

	return createdPersonDto, nil
}

func (s *PersonService) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*dtos.PersonDto, error) {
	log.Info().
		Str("person_id", personDto.Id.String()).
		Msg("Updating person")

	log.Debug().Str("person_id", personDto.Id.String()).Msg("Checking if person exists")
	_, err := s.GetPersonById(ctx, personDto.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personDto.Id.String()).
			Msg("Person not found for update")
		return nil, err
	}

	log.Debug().Str("person_id", personDto.Id.String()).Msg("Updating person in database")
	updatedPerson, err := s.personDriver.UpdatePerson(ctx, personDto)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personDto.Id.String()).
			Msg("Failed to update person in database")
		return nil, err
	}

	log.Info().
		Str("person_id", personDto.Id.String()).
		Str("name", updatedPerson.Name).
		Str("surname", updatedPerson.Surname).
		Uint32("age", updatedPerson.Age).
		Str("country", updatedPerson.Country).
		Str("gender", string(updatedPerson.Gender)).
		Msg("Person updated successfully")

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
	log.Info().
		Str("person_id", personId.String()).
		Msg("Deleting person")

	log.Debug().Str("person_id", personId.String()).Msg("Checking if person exists")
	person, err := s.GetPersonById(ctx, personId)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personId.String()).
			Msg("Person not found for deletion")
		return err
	}

	log.Debug().
		Str("person_id", personId.String()).
		Str("name", *person.Name).
		Str("surname", *person.Surname).
		Msg("Found person to delete")

	log.Debug().Str("person_id", personId.String()).Msg("Deleting person from database")
	err = s.personDriver.DeletePerson(ctx, personId)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personId.String()).
			Msg("Failed to delete person from database")
		return err
	}

	log.Info().
		Str("person_id", personId.String()).
		Msg("Person deleted successfully")

	return nil
}

func (s *PersonService) GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]dtos.PersonDto, error) {
	log.Info().Msg("Getting persons with filters")

	log.Debug().Msg("Validating filter parameters")
	if err := validateGetPersonDto(getPersonDto); err != nil {
		log.Warn().
			Err(err).
			Msg("Invalid filter parameters")
		return nil, err
	}

	log.Debug().Msg("Fetching persons from database")
	persons, err := s.personDriver.GetPersons(ctx, getPersonDto)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get persons from database")
		return nil, err
	}

	log.Debug().Int("count", len(persons)).Msg("Converting persons to DTOs")
	personDtos := make([]dtos.PersonDto, len(persons))
	for i, person := range persons {
		personDtos[i] = *mapPersonToDto(&person)
	}

	log.Info().
		Int("count", len(personDtos)).
		Msg("Persons retrieved successfully")

	return personDtos, nil
}

func (s *PersonService) GetPersonById(ctx context.Context, personId pgtype.UUID) (*dtos.PersonDto, error) {
	log.Info().
		Str("person_id", personId.String()).
		Msg("Getting person by ID")

	log.Debug().Str("person_id", personId.String()).Msg("Fetching person from database")
	person, err := s.personDriver.GetPersonById(ctx, personId)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personId.String()).
			Msg("Failed to get person from database")
		return nil, err
	}

	log.Info().
		Str("person_id", personId.String()).
		Str("name", person.Name).
		Str("surname", person.Surname).
		Uint32("age", person.Age).
		Str("country", person.Country).
		Str("gender", string(person.Gender)).
		Msg("Person retrieved successfully")

	personDto := mapPersonToDto(person)

	return personDto, nil
}

func getAge(name string) (uint32, error) {
	ageUrl := os.Getenv("AGE_URL") + name
	log.Debug().Str("url", ageUrl).Msg("Making request to age API")

	resp, err := http.Get(ageUrl)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", ageUrl).
			Msg(custom_errors.ErrHttpGet.Message)
		return 0, custom_errors.ErrHttpGet
	}
	defer resp.Body.Close()

	log.Debug().Int("status_code", resp.StatusCode).Msg("Age API response received")
	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", ageUrl).
			Msg(custom_errors.ErrGetAgeStatusCode.Message)
		return 0, custom_errors.ErrGetAgeStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", ageUrl).
			Msg(custom_errors.ErrGetAgeReadBody.Message)
		return 0, custom_errors.ErrGetAgeReadBody
	}

	var ageDto dtos.AgeDto
	if err = json.Unmarshal(body, &ageDto); err != nil {
		log.Error().
			Err(err).
			Str("body", string(body)).
			Msg(custom_errors.ErrGetAgeUnmarshalBody.Message)
		return 0, custom_errors.ErrGetAgeUnmarshalBody
	}

	log.Debug().
		Uint32("age", ageDto.Age).
		Str("name", name).
		Msg("Age successfully determined")

	return ageDto.Age, nil
}

func getGender(name string) (models.GenderType, error) {
	genderUrl := os.Getenv("GENDER_URL") + name
	log.Debug().Str("url", genderUrl).Msg("Making request to gender API")

	resp, err := http.Get(genderUrl)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", genderUrl).
			Msg(custom_errors.ErrHttpGet.Message)
		return "", custom_errors.ErrHttpGet
	}
	defer resp.Body.Close()

	log.Debug().Int("status_code", resp.StatusCode).Msg("Gender API response received")
	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", genderUrl).
			Msg(custom_errors.ErrGetGenderStatusCode.Message)
		return "", custom_errors.ErrGetGenderStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", genderUrl).
			Msg(custom_errors.ErrGetGenderReadBody.Message)
		return "", custom_errors.ErrGetGenderReadBody
	}

	var genderDto dtos.GenderDto
	if err = json.Unmarshal(body, &genderDto); err != nil {
		log.Error().
			Err(err).
			Str("body", string(body)).
			Msg(custom_errors.ErrGetGenderUnmarshalBody.Message)
		return "", custom_errors.ErrGetGenderUnmarshalBody
	}

	gender := models.GenderType(genderDto.Gender)
	log.Debug().
		Str("gender", string(gender)).
		Str("name", name).
		Msg("Gender determined")

	if gender != models.Male && gender != models.Female {
		log.Error().
			Str("gender", string(gender)).
			Str("name", name).
			Msg(custom_errors.ErrGotInvalidGender.Message)
		return "", custom_errors.ErrGotInvalidGender
	}

	log.Debug().
		Str("gender", string(gender)).
		Str("name", name).
		Msg("Gender successfully validated")

	return gender, nil
}

func getCountry(name string) (string, error) {
	countryUrl := os.Getenv("COUNTRY_URL") + name
	log.Debug().Str("url", countryUrl).Msg("Making request to country API")

	resp, err := http.Get(countryUrl)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", countryUrl).
			Msg(custom_errors.ErrHttpGet.Message)
		return "", custom_errors.ErrHttpGet
	}
	defer resp.Body.Close()

	log.Debug().Int("status_code", resp.StatusCode).Msg("Country API response received")
	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("url", countryUrl).
			Msg(custom_errors.ErrGetCountryStatusCode.Message)
		return "", custom_errors.ErrGetCountryStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", countryUrl).
			Msg(custom_errors.ErrGetCountryReadBody.Message)
		return "", custom_errors.ErrGetCountryReadBody
	}

	var countryDto dtos.CountryDto
	if err = json.Unmarshal(body, &countryDto); err != nil {
		log.Error().
			Err(err).
			Str("body", string(body)).
			Msg(custom_errors.ErrGetCountryUnmarshalBody.Message)
		return "", custom_errors.ErrGetCountryUnmarshalBody
	}

	log.Debug().
		Int("countries_count", len(countryDto.Countries)).
		Str("name", name).
		Msg("Country candidates determined")

	maxProb := countryDto.Countries[0]
	for _, country := range countryDto.Countries {
		if country.Probability > maxProb.Probability {
			maxProb = country
		}
	}

	log.Debug().
		Str("country", maxProb.Id).
		Float64("probability", maxProb.Probability).
		Str("name", name).
		Msg("Country successfully determined")

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
	log.Debug().Msg("Validating GetPersonDto")

	if len(getPersonDto.Ids) > 0 {
		for _, id := range getPersonDto.Ids {
			if !id.Valid {
				log.Error().
					Str("id", id.String()).
					Msg(custom_errors.ErrInvalidUuid.Message)
				return custom_errors.ErrInvalidUuid
			}
		}
	}

	if getPersonDto.Limit != nil && *getPersonDto.Limit < 0 {
		log.Error().
			Uint32("limit", *getPersonDto.Limit).
			Msg(custom_errors.ErrLimitValue.Message)
		return custom_errors.ErrLimitValue
	}

	if getPersonDto.Offset != nil && *getPersonDto.Offset < 0 {
		log.Error().
			Uint32("offset", *getPersonDto.Offset).
			Msg(custom_errors.ErrOffsetValue.Message)
		return custom_errors.ErrOffsetValue
	}

	if getPersonDto.LowAge != nil && *getPersonDto.LowAge < 0 {
		log.Error().
			Uint32("low_age", *getPersonDto.LowAge).
			Msg(custom_errors.ErrLowAgeValue.Message)
		return custom_errors.ErrLowAgeValue
	}

	if getPersonDto.HighAge != nil && *getPersonDto.HighAge < 0 {
		log.Error().
			Uint32("high_age", *getPersonDto.HighAge).
			Msg(custom_errors.ErrHighAgeValue.Message)
		return custom_errors.ErrHighAgeValue
	}

	if getPersonDto.Gender != nil && *getPersonDto.Gender != "male" && *getPersonDto.Gender != "female" {
		log.Error().
			Str("gender", *getPersonDto.Gender).
			Msg(custom_errors.ErrInvalidGender.Message)
		return custom_errors.ErrInvalidGender
	}

	return nil
}
