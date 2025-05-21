package services

import (
	"context"
	"effective-mobile/internal/drivers"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"effective-mobile/internal/models/custom_errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type MockPersonDriver struct {
	drivers.PersonDriver
	mock.Mock
}

func (m *MockPersonDriver) CreatePerson(ctx context.Context, person *models.Person) error {
	args := m.Called(ctx, person)
	return args.Error(0)
}

func (m *MockPersonDriver) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*models.Person, error) {
	args := m.Called(ctx, personDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Person), args.Error(1)
}

func (m *MockPersonDriver) DeletePerson(ctx context.Context, personId pgtype.UUID) error {
	args := m.Called(ctx, personId)
	return args.Error(0)
}

func (m *MockPersonDriver) GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]models.Person, error) {
	args := m.Called(ctx, getPersonDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Person), args.Error(1)
}

func (m *MockPersonDriver) GetPersonById(ctx context.Context, id pgtype.UUID) (*models.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Person), args.Error(1)
}

func setupMockServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	return server
}

func TestCreatePerson(t *testing.T) {
	ctx := context.Background()

	ageServer := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":3800,"name":"Dmitriy","age":44}`))
	})
	defer ageServer.Close()
	os.Setenv("AGE_URL", ageServer.URL+"/?name=")

	genderServer := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":34891,"name":"Dmitriy","gender":"male","probability":1.0}`))
	})
	defer genderServer.Close()
	os.Setenv("GENDER_URL", genderServer.URL+"/?name=")

	countryServer := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"count":1295,"name":"Dmitriy","country":[{"country_id":"UA","probability":0.3577828495179683},{"country_id":"RU","probability":0.1611119317856264},{"country_id":"KZ","probability":0.04676676792430468},{"country_id":"BY","probability":0.04444553279789973},{"country_id":"UZ","probability":0.018182263417322615}]}`))
	})
	defer countryServer.Close()
	os.Setenv("COUNTRY_URL", countryServer.URL+"/?name=")

	t.Run("CreatePerson without patronymic", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		createPersonDto := dtos.CreatePersonDto{
			Name:    "Ivan",
			Surname: "Ivanov",
		}

		mockDriver.On("CreatePerson", mock.Anything, mock.Anything).Return(nil)

		personDto, err := service.CreatePerson(ctx, createPersonDto)

		assert.NoError(t, err)
		assert.NotNil(t, personDto)
		assert.Equal(t, createPersonDto.Name, *personDto.Name)
		assert.Equal(t, createPersonDto.Surname, *personDto.Surname)
		assert.NotEmpty(t, *personDto.Age)
		assert.NotEmpty(t, *personDto.Gender)
		assert.NotEmpty(t, *personDto.Country)
		mockDriver.AssertExpectations(t)
	})

	t.Run("CreatePerson with patronymic", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		patronymic := "Ivanovich"
		createPersonDto := dtos.CreatePersonDto{
			Name:       "Ivan",
			Surname:    "Ivanov",
			Patronymic: &patronymic,
		}

		mockDriver.On("CreatePerson", mock.Anything, mock.Anything).Return(nil)

		personDto, err := service.CreatePerson(ctx, createPersonDto)

		assert.NoError(t, err)
		assert.NotNil(t, personDto)
		assert.Equal(t, createPersonDto.Name, *personDto.Name)
		assert.Equal(t, createPersonDto.Surname, *personDto.Surname)
		assert.NotEmpty(t, *personDto.Age)
		assert.NotEmpty(t, *personDto.Gender)
		assert.NotEmpty(t, *personDto.Country)
		mockDriver.AssertExpectations(t)
	})
}

func TestUpdatePerson(t *testing.T) {
	ctx := context.Background()

	t.Run("UpdatePerson with existing id", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		var age uint32 = 25
		country := "EN"
		updatePersonDto := dtos.PersonDto{
			Id:      id,
			Age:     &age,
			Country: &country,
		}

		mockDriver.On("GetPersonById", mock.Anything, mock.Anything).Return(
			&models.Person{
				Id:      id,
				Age:     20,
				Country: "RU",
			},
			nil,
		)
		mockDriver.On("UpdatePerson", mock.Anything, mock.Anything).Return(
			&models.Person{
				Id:      id,
				Age:     age,
				Country: country,
			},
			nil,
		)

		personDto, err := service.UpdatePerson(ctx, updatePersonDto)
		assert.NoError(t, err)
		assert.NotNil(t, personDto)
		assert.Equal(t, id, personDto.Id)
		assert.Equal(t, age, *personDto.Age)
		assert.Equal(t, country, *personDto.Country)
		mockDriver.AssertExpectations(t)
	})

	t.Run("UpdatePerson with non-existing id", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		var age uint32 = 25
		country := "EN"
		updatePersonDto := dtos.PersonDto{
			Id:      id,
			Age:     &age,
			Country: &country,
		}

		mockDriver.On("GetPersonById", mock.Anything, mock.Anything).Return(nil, custom_errors.ErrPersonNotFound)

		personDto, err := service.UpdatePerson(ctx, updatePersonDto)
		assert.Error(t, err)
		assert.Nil(t, personDto)
		assert.Equal(t, custom_errors.ErrPersonNotFound, err)
		mockDriver.AssertExpectations(t)
	})
}

func TestDeletePerson(t *testing.T) {
	ctx := context.Background()

	t.Run("DeletePerson with existing id", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}

		mockDriver.On("GetPersonById", mock.Anything, mock.Anything).Return(
			&models.Person{
				Id:      id,
				Age:     20,
				Country: "RU",
			},
			nil,
		)
		mockDriver.On("DeletePerson", mock.Anything, mock.Anything).Return(nil)

		err := service.DeletePerson(ctx, id)
		assert.NoError(t, err)
		mockDriver.AssertExpectations(t)
	})

	t.Run("DeletePerson with non-existing id", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}

		mockDriver.On("GetPersonById", mock.Anything, mock.Anything).Return(nil, custom_errors.ErrPersonNotFound)

		err := service.DeletePerson(ctx, id)
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrPersonNotFound, err)
		mockDriver.AssertExpectations(t)
	})
}

func TestGetPersons(t *testing.T) {
	ctx := context.Background()

	t.Run("GetPersons without filters", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		mockDriver.On("GetPersons", mock.Anything, mock.Anything).Return([]models.Person{}, nil)

		personDtos, err := service.GetPersons(ctx, dtos.GetPersonDto{})
		assert.NoError(t, err)
		assert.NotNil(t, personDtos)
		mockDriver.AssertExpectations(t)
	})

	t.Run("GetPersons with valid filters", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		var lowAge uint32 = 25
		getPersonDtos := dtos.GetPersonDto{
			Names:  []string{"Ivan", "Anna"},
			LowAge: &lowAge,
		}

		mockDriver.On("GetPersons", mock.Anything, mock.Anything).Return([]models.Person{}, nil)

		personDtos, err := service.GetPersons(ctx, getPersonDtos)
		assert.NoError(t, err)
		assert.NotNil(t, personDtos)
		mockDriver.AssertExpectations(t)
	})

	t.Run("GetPersons with invalid filters", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		gender := "non-binary"
		getPersonDtos := dtos.GetPersonDto{
			Names:  []string{"Ivan", "Anna"},
			Gender: &gender,
		}

		personDto, err := service.GetPersons(ctx, getPersonDtos)
		assert.Error(t, err)
		assert.Nil(t, personDto)
		assert.Equal(t, custom_errors.ErrInvalidGender, err)
		mockDriver.AssertExpectations(t)
	})
}

func TestGetPersonById(t *testing.T) {
	ctx := context.Background()

	t.Run("GetPersonById with existing id", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}

		mockDriver.On("GetPersonById", mock.Anything, mock.Anything).Return(&models.Person{}, nil)

		personDto, err := service.GetPersonById(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, personDto)
		mockDriver.AssertExpectations(t)
	})

	t.Run("GetPersonById with non-existing id", func(t *testing.T) {
		mockDriver := new(MockPersonDriver)
		service := NewPersonService(mockDriver)

		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}

		mockDriver.On("GetPersonById", mock.Anything, mock.Anything).Return(nil, custom_errors.ErrPersonNotFound)

		personDto, err := service.GetPersonById(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, personDto)
		assert.Equal(t, custom_errors.ErrPersonNotFound, err)
		mockDriver.AssertExpectations(t)
	})
}
