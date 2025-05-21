package api

import (
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models/custom_errors"
	"effective-mobile/internal/services"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"net/http"
)

// PersonHandler @title Person API
// @version 1.0
// @description API для управления данными о людях
// @BasePath /api/v1
type PersonHandler struct {
	personService services.PersonServiceInterface
}

func NewPersonHandler(personService services.PersonServiceInterface) *PersonHandler {
	log.Debug().Msg("Initializing PersonHandler")
	return &PersonHandler{personService: personService}
}

// CreatePerson godoc
// @Summary Создание новой записи о человеке
// @Description Создаёт новую запись о человеке на основе переданных данных
// @Tags persons
// @Accept json
// @Produce json
// @Param person body dtos.CreatePersonDto true "Информация о человеке"
// @Success 201 {object} dtos.PersonDto "Созданная запись о человеке"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /persons [post]
func (h *PersonHandler) CreatePerson(c *gin.Context) {
	log.Info().Msg("CreatePerson handler started")
	reqId := getRequestID(c)

	var createPersonDto dtos.CreatePersonDto
	if err := c.ShouldBindJSON(&createPersonDto); err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Str("payload", c.Request.URL.String()).
			Msg(custom_errors.ErrBindJsonBody.Message)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to create person: " + err.Error()})
		return
	}

	log.Debug().
		Str("request_id", reqId).
		Str("name", createPersonDto.Name).
		Str("surname", createPersonDto.Surname).
		Msg("Attempting to create person")

	personDto, err := h.personService.CreatePerson(c.Request.Context(), createPersonDto)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		log.Warn().
			Err(err).
			Str("request_id", reqId).
			Str("error_type", "user_error").
			Msg("User error when creating person")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to create person: " + userErr.Error()})
		return
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Msg("Server error when creating person")
		c.JSON(http.StatusInternalServerError, gin.H{"CreatePerson error": err.Error()})
		return
	}

	log.Info().
		Str("request_id", reqId).
		Str("person_id", personDto.Id.String()).
		Msg("Person created successfully")

	c.JSON(http.StatusCreated, personDto)
}

// UpdatePerson godoc
// @Summary Обновление данных о человеке
// @Description Обновляет существующую запись о человеке на основе переданных данных
// @Tags persons
// @Accept json
// @Produce json
// @Param person body dtos.PersonDto true "Информация о человеке для обновления"
// @Success 200 {object} dtos.PersonDto "Обновленная запись о человеке"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /persons [put]
func (h *PersonHandler) UpdatePerson(c *gin.Context) {
	log.Info().Msg("UpdatePerson handler started")
	reqId := getRequestID(c)

	var updatePersonDto dtos.PersonDto
	if err := c.ShouldBindJSON(&updatePersonDto); err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Str("payload", c.Request.URL.String()).
			Msg(custom_errors.ErrBindJsonBody.Message)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to update person: " + err.Error()})
		return
	}

	log.Debug().
		Str("request_id", reqId).
		Str("person_id", updatePersonDto.Id.String()).
		Msg("Attempting to update person")

	personDto, err := h.personService.UpdatePerson(c.Request.Context(), updatePersonDto)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		log.Warn().
			Err(err).
			Str("request_id", reqId).
			Str("person_id", updatePersonDto.Id.String()).
			Str("error_type", "user_error").
			Msg("User error when updating person")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to update person: " + userErr.Error()})
		return
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Str("person_id", updatePersonDto.Id.String()).
			Msg("Server error when updating person")
		c.JSON(http.StatusInternalServerError, gin.H{"UpdatePerson error": err.Error()})
		return
	}

	log.Info().
		Str("request_id", reqId).
		Str("person_id", personDto.Id.String()).
		Msg("Person updated successfully")

	c.JSON(http.StatusOK, personDto)
}

// DeletePerson godoc
// @Summary Удаление записи о человеке
// @Description Удаляет запись о человеке по указанному ID
// @Tags persons
// @Produce json
// @Param id path string true "ID человека" format(uuid)
// @Success 200 {object} map[string]string "Сообщение об успешном удалении"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /persons/{id} [delete]
func (h *PersonHandler) DeletePerson(c *gin.Context, personId pgtype.UUID) {
	log.Info().Msg("DeletePerson handler started")
	reqId := getRequestID(c)

	log.Debug().
		Str("request_id", reqId).
		Str("person_id", personId.String()).
		Msg("Attempting to delete person")

	err := h.personService.DeletePerson(c.Request.Context(), personId)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		log.Warn().
			Err(err).
			Str("request_id", reqId).
			Str("person_id", personId.String()).
			Str("error_type", "user_error").
			Msg("User error when deleting person")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to delete person: " + userErr.Error()})
		return
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Str("person_id", personId.String()).
			Msg("Server error when deleting person")
		c.JSON(http.StatusInternalServerError, gin.H{"DeletePerson error": err.Error()})
		return
	}

	log.Info().
		Str("request_id", reqId).
		Str("person_id", personId.String()).
		Msg("Person deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "person " + personId.String() + " deleted successfully!"})
}

// GetPersons godoc
// @Summary Получение данных о нескольких людях
// @Description Возвращает список людей согласно указанным фильтрам
// @Tags persons
// @Accept json
// @Produce json
// @Param filter body dtos.GetPersonDto true "Параметры фильтрации"
// @Success 200 {array} dtos.PersonDto "Список людей"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /persons [get]
func (h *PersonHandler) GetPersons(c *gin.Context) {
	log.Info().Msg("GetPersons handler started")
	reqId := getRequestID(c)

	var getPersonsDto dtos.GetPersonDto
	if err := c.ShouldBindJSON(&getPersonsDto); err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Str("payload", c.Request.URL.String()).
			Msg(custom_errors.ErrBindJsonBody.Message)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to get persons: " + err.Error()})
		return
	}

	log.Debug().
		Str("request_id", reqId).
		Int("ids_count", len(getPersonsDto.Ids)).
		Msg("Attempting to get persons")

	personDtos, err := h.personService.GetPersons(c.Request.Context(), getPersonsDto)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		log.Warn().
			Err(err).
			Str("request_id", reqId).
			Str("error_type", "user_error").
			Msg("User error when getting persons")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to get persons: " + userErr.Error()})
		return
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Msg("Server error when getting persons")
		c.JSON(http.StatusInternalServerError, gin.H{"GetPersons error": err.Error()})
		return
	}

	log.Info().
		Str("request_id", reqId).
		Int("found_count", len(personDtos)).
		Msg("Persons retrieved successfully")

	c.JSON(http.StatusOK, personDtos)
}

// GetPersonById godoc
// @Summary Получение данных о человеке по ID
// @Description Возвращает информацию о человеке по указанному ID
// @Tags persons
// @Produce json
// @Param id path string true "ID человека" format(uuid)
// @Success 200 {object} dtos.PersonDto "Информация о человеке"
// @Failure 400 {object} map[string]string "Ошибка валидации запроса"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /persons/{id} [get]
func (h *PersonHandler) GetPersonById(c *gin.Context, personId pgtype.UUID) {
	log.Info().Msg("GetPersonById handler started")
	reqId := getRequestID(c)

	log.Debug().
		Str("request_id", reqId).
		Str("person_id", personId.String()).
		Msg("Attempting to get person by ID")

	personDto, err := h.personService.GetPersonById(c.Request.Context(), personId)
	var userErr *custom_errors.UserError
	if errors.As(err, &userErr) {
		log.Warn().
			Err(err).
			Str("request_id", reqId).
			Str("person_id", personId.String()).
			Str("error_type", "user_error").
			Msg("User error when getting person by ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format to get person by id: " + userErr.Error()})
		return
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", reqId).
			Str("person_id", personId.String()).
			Msg("Server error when getting person by ID")
		c.JSON(http.StatusInternalServerError, gin.H{"GetPersonById error": err.Error()})
		return
	}

	log.Info().
		Str("request_id", reqId).
		Str("person_id", personId.String()).
		Msg("Person retrieved successfully")

	c.JSON(http.StatusOK, personDto)
}

func getRequestID(c *gin.Context) string {
	reqID, exists := c.Get("RequestID")
	if !exists {
		return "unknown"
	}
	return reqID.(string)
}
