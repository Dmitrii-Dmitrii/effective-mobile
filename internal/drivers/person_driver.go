package drivers

import (
	"context"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
)

type PersonDriver struct {
	adapter Adapter
}

func NewPersonDriver(adapter Adapter) *PersonDriver {
	return &PersonDriver{adapter: adapter}
}

func (d *PersonDriver) CreatePerson(ctx context.Context, person *models.Person) error {
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

	return err
}

func (d *PersonDriver) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) (*models.Person, error) {
	query := `UPDATE persons`

	setValues, args, argCnt := setArgumentsForUpdate(personDto)

	if len(setValues) == 0 {
		return nil, errors.New("no fields to update")
	}

	args = append(args, personDto.Id)

	query += fmt.Sprintf(" SET %s WHERE id = $%d",
		strings.Join(setValues, ", "),
		argCnt,
	)

	_, err := d.adapter.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	person, err := d.GetPersonById(ctx, personDto.Id)
	if err != nil {
		return nil, err
	}

	return person, nil
}

func (d *PersonDriver) DeletePerson(ctx context.Context, personId pgtype.UUID) error {
	_, err := d.adapter.Exec(ctx, queryDeletePerson, personId)
	return err
}

func (d *PersonDriver) GetPersons(ctx context.Context, getPersonDto dtos.GetPersonDto) ([]models.Person, error) {
	var persons []models.Person
	query := queryGetPersons

	setValues, args, argCnt := setArgumentsForGet(getPersonDto)

	if len(setValues) > 0 {
		query += " WHERE " + strings.Join(setValues, " AND ")
	}

	query += " ORDER BY id"

	if getPersonDto.Limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCnt)
		args = append(args, *getPersonDto.Limit)
		argCnt++
	}

	if getPersonDto.Offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", argCnt)
		args = append(args, *getPersonDto.Offset)
		argCnt++
	}

	rows, err := d.adapter.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		person := models.Person{}

		err := rows.Scan(
			&person.Id,
			&person.Name,
			&person.Surname,
			&person.Patronymic,
			&person.Age,
			&person.Gender,
			&person.Country,
		)
		if err != nil {
			return nil, err
		}

		persons = append(persons, person)
	}

	return persons, nil
}

func (d *PersonDriver) GetPersonById(ctx context.Context, id pgtype.UUID) (*models.Person, error) {
	person := models.Person{Id: id}

	err := d.adapter.QueryRow(ctx, queryGetPersonById, id).Scan(
		&person.Name,
		&person.Surname,
		&person.Patronymic,
		&person.Age,
		&person.Gender,
		&person.Country,
	)
	if err != nil {
		return nil, err
	}

	return &person, nil
}

func setArgumentsForUpdate(person dtos.PersonDto) ([]string, []interface{}, int) {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argCnt := 1

	if person.Name != nil {
		setValues = append(setValues, fmt.Sprintf("name = $%d", argCnt))
		args = append(args, *person.Name)
		argCnt++
	}

	if person.Surname != nil {
		setValues = append(setValues, fmt.Sprintf("surname = $%d", argCnt))
		args = append(args, *person.Surname)
		argCnt++
	}

	if person.Patronymic != nil {
		setValues = append(setValues, fmt.Sprintf("patronymic = $%d", argCnt))
		args = append(args, *person.Patronymic)
		argCnt++
	}

	if person.Age != nil {
		setValues = append(setValues, fmt.Sprintf("age = $%d", argCnt))
		args = append(args, *person.Age)
		argCnt++
	}

	if person.Gender != nil {
		setValues = append(setValues, fmt.Sprintf("gender = $%d", argCnt))
		args = append(args, *person.Gender)
		argCnt++
	}

	if person.Country != nil {
		setValues = append(setValues, fmt.Sprintf("country = $%d", argCnt))
		args = append(args, *person.Country)
		argCnt++
	}

	return setValues, args, argCnt
}

func setArgumentsForGet(getPersonDto dtos.GetPersonDto) ([]string, []interface{}, int) {
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
	}

	if len(getPersonDto.Names) > 0 {
		names := make([]string, 0, len(getPersonDto.Names))
		for _, name := range getPersonDto.Names {
			names = append(names, fmt.Sprintf("$%d", argCnt))
			args = append(args, name)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("name IN (%s)", strings.Join(names, ", ")))
	}

	if len(getPersonDto.Surnames) > 0 {
		surnames := make([]string, 0, len(getPersonDto.Surnames))
		for _, surname := range getPersonDto.Surnames {
			surnames = append(surnames, fmt.Sprintf("$%d", argCnt))
			args = append(args, surname)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("surname IN (%s)", strings.Join(surnames, ", ")))
	}

	if len(getPersonDto.Patronymics) > 0 {
		patronymics := make([]string, 0, len(getPersonDto.Patronymics))
		for _, patronymic := range getPersonDto.Patronymics {
			patronymics = append(patronymics, fmt.Sprintf("$%d", argCnt))
			args = append(args, patronymic)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("patronymic IN (%s)", strings.Join(patronymics, ", ")))
	}

	if getPersonDto.LowAge != nil {
		setValues = append(setValues, fmt.Sprintf("age >= $%d", argCnt))
		args = append(args, *getPersonDto.LowAge)
		argCnt++
	}

	if getPersonDto.HighAge != nil {
		setValues = append(setValues, fmt.Sprintf("age <= $%d", argCnt))
		args = append(args, *getPersonDto.HighAge)
		argCnt++
	}

	if getPersonDto.Gender != nil {
		setValues = append(setValues, fmt.Sprintf("gender = $%d", argCnt))
		args = append(args, *getPersonDto.Gender)
		argCnt++
	}

	if len(getPersonDto.Countries) > 0 {
		countries := make([]string, 0, len(getPersonDto.Countries))
		for _, country := range getPersonDto.Countries {
			countries = append(countries, fmt.Sprintf("$%d", argCnt))
			args = append(args, country)
			argCnt++
		}
		setValues = append(setValues, fmt.Sprintf("country IN (%s)", strings.Join(countries, ", ")))
	}

	return setValues, args, argCnt
}
