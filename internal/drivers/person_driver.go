package drivers

import (
	"context"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"effective-mobile/internal/models/custom_errors"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"strings"
)

type PersonDriver struct {
	adapter Adapter
}

func NewPersonDriver(adapter Adapter) *PersonDriver {
	log.Debug().Msg("Initializing PersonDriver")
	return &PersonDriver{adapter: adapter}
}

func (d *PersonDriver) CreatePerson(ctx context.Context, person *models.Person) error {
	log.Info().
		Str("person_id", person.Id.String()).
		Str("name", person.Name).
		Str("surname", person.Surname).
		Uint32("age", person.Age).
		Str("gender", string(person.Gender)).
		Str("country", person.Country).
		Msg("Creating person in database")

	_, err := d.adapter.Exec(
		ctx,
		queryCreatePerson,
		person.Id,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Country,
	)

	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", person.Id.String()).
			Str("name", person.Name).
			Str("surname", person.Surname).
			Msg(custom_errors.ErrCreatePerson.Message)
		return custom_errors.ErrCreatePerson
	}

	log.Debug().
		Str("person_id", person.Id.String()).
		Msg("Person successfully created in database")

	return nil
}

func (d *PersonDriver) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*models.Person, error) {
	log.Info().
		Str("person_id", personDto.Id.String()).
		Msg("Updating person in database")

	query := `UPDATE persons`

	setValues, args, argCnt := setArgumentsForUpdate(personDto)

	if len(setValues) == 0 {
		log.Error().
			Str("person_id", personDto.Id.String()).
			Msg(custom_errors.ErrNoFieldsToUpdate.Message)
		return nil, custom_errors.ErrNoFieldsToUpdate
	}

	log.Debug().
		Str("person_id", personDto.Id.String()).
		Int("fields_to_update", len(setValues)).
		Msg("Prepared update query arguments")

	args = append(args, personDto.Id)

	query += fmt.Sprintf(" SET %s WHERE id = $%d",
		strings.Join(setValues, ", "),
		argCnt,
	)

	log.Debug().
		Str("person_id", personDto.Id.String()).
		Str("query", query).
		Msg("Executing update query")

	_, err := d.adapter.Exec(ctx, query, args...)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personDto.Id.String()).
			Str("query", query).
			Msg(custom_errors.ErrUpdatePerson.Message)
		return nil, custom_errors.ErrUpdatePerson
	}

	log.Debug().
		Str("person_id", personDto.Id.String()).
		Msg("Fetching updated person")

	person, err := d.GetPersonById(ctx, personDto.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personDto.Id.String()).
			Msg("Failed to fetch updated person")
		return nil, err
	}

	log.Debug().
		Str("person_id", personDto.Id.String()).
		Str("name", person.Name).
		Str("surname", person.Surname).
		Msg("Successfully updated person in database")

	return person, nil
}

func (d *PersonDriver) DeletePerson(ctx context.Context, personId pgtype.UUID) error {
	log.Info().
		Str("person_id", personId.String()).
		Msg("Deleting person from database")

	_, err := d.adapter.Exec(ctx, queryDeletePerson, personId)
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", personId.String()).
			Msg(custom_errors.ErrDeletePerson.Message)
		return custom_errors.ErrDeletePerson
	}

	log.Debug().
		Str("person_id", personId.String()).
		Msg("Successfully deleted person from database")

	return nil
}

func (d *PersonDriver) GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]models.Person, error) {
	log.Info().Msg("Fetching persons from database with filters")

	var persons []models.Person
	query := queryGetPersons

	setValues, args, argCnt := setArgumentsForGet(getPersonDto)

	log.Debug().
		Int("filter_conditions", len(setValues)).
		Int("arguments_count", len(args)).
		Msg("Prepared query filters")

	if len(setValues) > 0 {
		query += " WHERE " + strings.Join(setValues, " AND ")
	}

	query += " ORDER BY id"

	if getPersonDto.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCnt)
		args = append(args, *getPersonDto.Limit)
		argCnt++
		log.Debug().
			Uint32("limit", *getPersonDto.Limit).
			Msg("Applied limit to query")
	}

	if getPersonDto.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argCnt)
		args = append(args, *getPersonDto.Offset)
		argCnt++
		log.Debug().
			Uint32("offset", *getPersonDto.Offset).
			Msg("Applied offset to query")
	}

	log.Debug().
		Str("query", query).
		Msg("Executing persons query")

	rows, err := d.adapter.Query(ctx, query, args...)
	if err != nil {
		log.Error().
			Err(err).
			Str("query", query).
			Msg(custom_errors.ErrGetPerson.Message)
		return nil, custom_errors.ErrGetPerson
	}
	defer rows.Close()

	personCount := 0
	for rows.Next() {
		person := models.Person{}

		err = rows.Scan(
			&person.Id,
			&person.Name,
			&person.Surname,
			&person.Patronymic,
			&person.Age,
			&person.Gender,
			&person.Country,
		)
		if err != nil {
			log.Error().
				Err(err).
				Msg(custom_errors.ErrScanRow.Message)
			return nil, custom_errors.ErrScanRow
		}

		persons = append(persons, person)
		personCount++
	}

	log.Debug().
		Int("found_count", personCount).
		Msg("Successfully fetched persons from database")

	return persons, nil
}

func (d *PersonDriver) GetPersonById(ctx context.Context, id pgtype.UUID) (*models.Person, error) {
	log.Info().
		Str("person_id", id.String()).
		Msg("Fetching person by ID from database")

	person := models.Person{Id: id}

	err := d.adapter.QueryRow(ctx, queryGetPersonById, id).Scan(
		&person.Name,
		&person.Surname,
		&person.Patronymic,
		&person.Age,
		&person.Gender,
		&person.Country,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Error().
			Err(err).
			Str("person_id", id.String()).
			Msg(custom_errors.ErrPersonNotFound.Message)
		return nil, custom_errors.ErrPersonNotFound
	}
	if err != nil {
		log.Error().
			Err(err).
			Str("person_id", id.String()).
			Msg(custom_errors.ErrGetPersonById.Message)
		return nil, custom_errors.ErrGetPersonById
	}

	log.Debug().
		Str("person_id", id.String()).
		Str("name", person.Name).
		Str("surname", person.Surname).
		Uint32("age", person.Age).
		Str("gender", string(person.Gender)).
		Str("country", person.Country).
		Msg("Successfully fetched person from database")

	return &person, nil
}

func setArgumentsForUpdate(person dtos.PersonDto) ([]string, []interface{}, int) {
	log.Debug().
		Str("person_id", person.Id.String()).
		Msg("Setting arguments for person update")

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argCnt := 1

	if person.Name != nil {
		setValues = append(setValues, fmt.Sprintf("name = $%d", argCnt))
		args = append(args, *person.Name)
		argCnt++
		log.Debug().
			Str("person_id", person.Id.String()).
			Str("name", *person.Name).
			Msg("Adding name to update fields")
	}

	if person.Surname != nil {
		setValues = append(setValues, fmt.Sprintf("surname = $%d", argCnt))
		args = append(args, *person.Surname)
		argCnt++
		log.Debug().
			Str("person_id", person.Id.String()).
			Str("surname", *person.Surname).
			Msg("Adding surname to update fields")
	}

	if person.Patronymic != nil {
		setValues = append(setValues, fmt.Sprintf("patronymic = $%d", argCnt))
		args = append(args, *person.Patronymic)
		argCnt++
		log.Debug().
			Str("person_id", person.Id.String()).
			Str("patronymic", *person.Patronymic).
			Msg("Adding patronymic to update fields")
	}

	if person.Age != nil {
		setValues = append(setValues, fmt.Sprintf("age = $%d", argCnt))
		args = append(args, *person.Age)
		argCnt++
		log.Debug().
			Str("person_id", person.Id.String()).
			Uint32("age", *person.Age).
			Msg("Adding age to update fields")
	}

	if person.Gender != nil {
		setValues = append(setValues, fmt.Sprintf("gender = $%d", argCnt))
		args = append(args, *person.Gender)
		argCnt++
		log.Debug().
			Str("person_id", person.Id.String()).
			Str("gender", *person.Gender).
			Msg("Adding gender to update fields")
	}

	if person.Country != nil {
		setValues = append(setValues, fmt.Sprintf("country = $%d", argCnt))
		args = append(args, *person.Country)
		argCnt++
		log.Debug().
			Str("person_id", person.Id.String()).
			Str("country", *person.Country).
			Msg("Adding country to update fields")
	}

	log.Debug().
		Str("person_id", person.Id.String()).
		Int("update_fields_count", len(setValues)).
		Msg("Prepared update arguments")

	return setValues, args, argCnt
}

func setArgumentsForGet(getPersonDto dtos.GetPersonDto) ([]string, []interface{}, int) {
	log.Debug().Msg("Setting arguments for persons query")

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argCnt := 1

	if len(getPersonDto.Ids) > 0 {
		ids := make([]string, 0, len(getPersonDto.Ids))
		for _, id := range getPersonDto.Ids {
			ids = append(ids, fmt.Sprintf("$%d", argCnt))
			args = append(args, id)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("id IN (%s)", strings.Join(ids, ", ")))
		log.Debug().
			Int("ids_count", len(getPersonDto.Ids)).
			Msg("Adding IDs filter to query")
	}

	if len(getPersonDto.Names) > 0 {
		names := make([]string, 0, len(getPersonDto.Names))
		for _, name := range getPersonDto.Names {
			names = append(names, fmt.Sprintf("$%d", argCnt))
			args = append(args, name)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("name IN (%s)", strings.Join(names, ", ")))
		log.Debug().
			Int("names_count", len(getPersonDto.Names)).
			Msg("Adding names filter to query")
	}

	if len(getPersonDto.Surnames) > 0 {
		surnames := make([]string, 0, len(getPersonDto.Surnames))
		for _, surname := range getPersonDto.Surnames {
			surnames = append(surnames, fmt.Sprintf("$%d", argCnt))
			args = append(args, surname)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("surname IN (%s)", strings.Join(surnames, ", ")))
		log.Debug().
			Int("surnames_count", len(getPersonDto.Surnames)).
			Msg("Adding surnames filter to query")
	}

	if len(getPersonDto.Patronymics) > 0 {
		patronymics := make([]string, 0, len(getPersonDto.Patronymics))
		for _, patronymic := range getPersonDto.Patronymics {
			patronymics = append(patronymics, fmt.Sprintf("$%d", argCnt))
			args = append(args, patronymic)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("patronymic IN (%s)", strings.Join(patronymics, ", ")))
		log.Debug().
			Int("patronymics_count", len(getPersonDto.Patronymics)).
			Msg("Adding patronymics filter to query")
	}

	if getPersonDto.LowAge != nil {
		setValues = append(setValues, fmt.Sprintf("age >= $%d", argCnt))
		args = append(args, *getPersonDto.LowAge)
		argCnt++
		log.Debug().
			Uint32("low_age", *getPersonDto.LowAge).
			Msg("Adding minimum age filter to query")
	}

	if getPersonDto.HighAge != nil {
		setValues = append(setValues, fmt.Sprintf("age <= $%d", argCnt))
		args = append(args, *getPersonDto.HighAge)
		argCnt++
		log.Debug().
			Uint32("high_age", *getPersonDto.HighAge).
			Msg("Adding maximum age filter to query")
	}

	if getPersonDto.Gender != nil {
		setValues = append(setValues, fmt.Sprintf("gender = $%d", argCnt))
		args = append(args, *getPersonDto.Gender)
		argCnt++
		log.Debug().
			Str("gender", *getPersonDto.Gender).
			Msg("Adding gender filter to query")
	}

	if len(getPersonDto.Countries) > 0 {
		countries := make([]string, 0, len(getPersonDto.Countries))
		for _, country := range getPersonDto.Countries {
			countries = append(countries, fmt.Sprintf("$%d", argCnt))
			args = append(args, country)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("country IN (%s)", strings.Join(countries, ", ")))
		log.Debug().
			Int("countries_count", len(getPersonDto.Countries)).
			Msg("Adding countries filter to query")
	}

	log.Debug().
		Int("filter_conditions", len(setValues)).
		Int("args_count", len(args)).
		Msg("Prepared query filter arguments")

	return setValues, args, argCnt
}
