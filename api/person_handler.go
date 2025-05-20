package api

import (
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
)

type PersonHandler struct {
	personService services.PersonService
}

func NewPersonHandler(personService services.PersonService) *PersonHandler {
	return &PersonHandler{personService: personService}
}

func (h *PersonHandler) CreatePerson(c *gin.Context) {
	var createPersonDto dtos.CreatePersonDto
	if err := c.ShouldBindJSON(&createPersonDto); err != nil {
		//log.Error().Err(err).Msg("failed to bind json body")
		//c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create pvz: " + err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personDto, err := h.personService.CreatePerson(c.Request.Context(), createPersonDto)
	//var userErr *custom_errors.UserError
	//if errors.As(err, &userErr) {
	//	c.JSON(http.StatusBadRequest, generated.Error{Message: "Invalid request format to create pvz: " + userErr.Error()})
	//	return
	//}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, personDto)
}

func (h *PersonHandler) UpdatePerson(c *gin.Context) {
	var updatePersonDto dtos.PersonDto
	if err := c.ShouldBindJSON(&updatePersonDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personDto, err := h.personService.UpdatePerson(c.Request.Context(), updatePersonDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, personDto)
}

func (h *PersonHandler) DeletePerson(c *gin.Context, personId pgtype.UUID) {
	if err := h.personService.DeletePerson(c.Request.Context(), personId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "person " + personId.String() + " deleted successfully!"})
}

func (h *PersonHandler) GetPersons(c *gin.Context) {
	var getPersonsDto dtos.GetPersonDto
	if err := c.ShouldBindJSON(&getPersonsDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	personDtos, err := h.personService.GetPersons(c.Request.Context(), getPersonsDto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, personDtos)
}

func (h *PersonHandler) GetPersonById(c *gin.Context, personId pgtype.UUID) {
	personDto, err := h.personService.GetPersonById(c.Request.Context(), personId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, personDto)
}
