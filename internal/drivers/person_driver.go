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

func (d *PersonDriver) UpdatePerson(ctx context.Context, personDto dtos.PersonDto) error {
	if personDto.Id == nil {
		return errors.New("person don't have id")
	}

	query := `UPDATE persons`

	setValues, args, argCnt := setArguments(personDto)

	if len(setValues) == 0 {
		return errors.New("no fields to update")
	}

	args = append(args, personDto.Id)

	query += fmt.Sprintf("SET %s WHERE id = $%d",
		strings.Join(setValues, ", "),
		argCnt,
	)

	_, err := d.adapter.Exec(ctx, query, args...)
	return err
}

func (d *PersonDriver) DeletePerson(ctx context.Context, id pgtype.UUID) error {
	_, err := d.adapter.Exec(ctx, queryDeletePerson, id)
	return err
}

func (d *PersonDriver) GetPerson(ctx context.Context, personDto dtos.PersonDto, limit, offset uint32) ([]models.Person, error) {
	var persons []models.Person
	query := queryGetPersons

	setValues, args, argCnt := setArguments(personDto)

	if len(setValues) == 0 {
		return nil, errors.New("no fields to update")
	}

	query += fmt.Sprintf("WHERE %s",
		strings.Join(setValues, "AND "),
	)

	query += fmt.Sprintf(" LIMIT $%d", argCnt)
	args = append(args, limit)
	argCnt++
	query += fmt.Sprintf(" OFFSET $%d", argCnt)
	args = append(args, offset)

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

func setArguments(person dtos.PersonDto) ([]string, []interface{}, int) {
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
