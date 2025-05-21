package drivers

const (
	queryCreatePerson = `
	INSERT INTO persons (id, name, surname, patronymic, age, gender, country)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	queryDeletePerson = `
	DELETE FROM persons 
	WHERE id = $1
`
	queryGetPersons = `
	SELECT id, name, surname, patronymic, age, gender, country
	FROM persons
`
	queryGetPersonById = `
	SELECT name, surname, patronymic, age, gender, country
	FROM persons
	WHERE id = $1
`
	createTestSchema = `
	CREATE TYPE gender_type AS ENUM (
		'male',
		'female'
		);
	
	CREATE TABLE IF NOT EXISTS persons
	(
		id UUID PRIMARY KEY,
		name TEXT NOT NULL,
		surname TEXT NOT NULL,
		patronymic TEXT NOT NULL,
		age INTEGER NOT NULL,
		gender gender_type NOT NULL,
		country TEXT
	);
`
)
