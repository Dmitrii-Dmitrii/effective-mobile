package api

import (
	"bytes"
	"context"
	"effective-mobile/internal/dtos"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockPersonService struct {
	mock.Mock
}

func (m *MockPersonService) CreatePerson(ctx context.Context, personDto dtos.CreatePersonDto) (*dtos.PersonDto, error) {
	args := m.Called(ctx, personDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.PersonDto), args.Error(1)
}

func (m *MockPersonService) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*dtos.PersonDto, error) {
	args := m.Called(ctx, personDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.PersonDto), args.Error(1)
}

func (m *MockPersonService) DeletePerson(ctx context.Context, personId pgtype.UUID) error {
	args := m.Called(ctx, personId)
	return args.Error(0)
}

func (m *MockPersonService) GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]dtos.PersonDto, error) {
	args := m.Called(ctx, getPersonDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dtos.PersonDto), args.Error(1)
}

func (m *MockPersonService) GetPersonById(ctx context.Context, personId pgtype.UUID) (*dtos.PersonDto, error) {
	args := m.Called(ctx, personId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.PersonDto), args.Error(1)
}

func TestCreatePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockPersonService)
	handler := NewPersonHandler(mockService)

	name := "name"
	surname := "surname"
	createPersonDto := dtos.CreatePersonDto{
		Name:    name,
		Surname: surname,
	}
	jsonData, _ := json.Marshal(createPersonDto)
	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}

	mockService.On("CreatePerson", mock.Anything, createPersonDto).Return(&dtos.PersonDto{
		Id:      id,
		Name:    &name,
		Surname: &surname,
	}, nil).Once()

	req, _ := http.NewRequest("POST", "/persons", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.POST("/persons", handler.CreatePerson)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	var response dtos.PersonDto
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, name, *response.Name)
	assert.Equal(t, surname, *response.Surname)
}

func TestUpdatePersons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockPersonService)
	handler := NewPersonHandler(mockService)

	country := "country"
	updatePersonDto := dtos.PersonDto{
		Country: &country,
	}
	jsonData, _ := json.Marshal(updatePersonDto)
	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}

	mockService.On("UpdatePerson", mock.Anything, updatePersonDto).Return(&dtos.PersonDto{
		Id:      id,
		Country: &country,
	}, nil).Once()

	req, _ := http.NewRequest("PUT", "/persons", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.PUT("/persons", handler.UpdatePerson)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response dtos.PersonDto
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, country, *response.Country)
}

func TestDeletePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockPersonService)
	handler := NewPersonHandler(mockService)

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}

	mockService.On("DeletePerson", mock.Anything, id).Return(nil).Once()

	router.DELETE("/persons/"+id.String(), func(c *gin.Context) {
		handler.DeletePerson(c, id)
	})

	req, _ := http.NewRequest("DELETE", "/persons/"+id.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response gin.H
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["message"], id.String())
}

func TestGetPersons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockPersonService)
	handler := NewPersonHandler(mockService)

	country := "country"
	getPersonsDto := dtos.GetPersonDto{
		Names: []string{country},
	}
	jsonData, _ := json.Marshal(getPersonsDto)
	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}

	mockService.On("GetPersons", mock.Anything, getPersonsDto).Return([]dtos.PersonDto{
		{
			Id:      id,
			Country: &country,
		},
	}, nil).Once()

	req, _ := http.NewRequest("GET", "/persons", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.GET("/persons", handler.GetPersons)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response []dtos.PersonDto
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, country, *response[0].Country)
}

func TestGetPersonById(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockPersonService)
	handler := NewPersonHandler(mockService)

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}

	mockService.On("GetPersonById", mock.Anything, id).Return(&dtos.PersonDto{
		Id: id,
	}, nil).Once()

	router.GET("/persons/"+id.String(), func(c *gin.Context) {
		handler.GetPersonById(c, id)
	})

	req, _ := http.NewRequest("GET", "/persons/"+id.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	var response dtos.PersonDto
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, id, response.Id)
}
