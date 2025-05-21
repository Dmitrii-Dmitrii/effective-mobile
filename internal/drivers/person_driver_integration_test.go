package drivers

import (
	"context"
	"effective-mobile/internal/dtos"
	"effective-mobile/internal/models"
	"effective-mobile/internal/models/custom_errors"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"slices"
	"strconv"
	"testing"
	"time"
)

func setupPostgresContainer(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	pgPort := "5432/tcp"
	dbName := "testdb"
	dbUser := "postgres"
	dbPassword := "postgres"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{pgPort},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(2 * time.Minute),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	hostIP, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := postgresContainer.MappedPort(ctx, nat.Port(pgPort))
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbUser, dbPassword, hostIP, mappedPort.Port(), dbName)

	time.Sleep(2 * time.Second)

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Could not connect to database: %v", err)
	}

	err = setupSchema(ctx, pool)
	if err != nil {
		pool.Close()
		postgresContainer.Terminate(ctx)
		t.Fatalf("Could not set up schema: %v", err)
	}

	return pool, func() {
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %s", err)
		}
	}
}

func setupSchema(ctx context.Context, pool *pgxpool.Pool) error {
	schema := createTestSchema

	_, err := pool.Exec(ctx, schema)
	return err
}

func createTestData(ctx context.Context, pool *pgxpool.Pool) ([]pgtype.UUID, error) {
	personIds := make([]pgtype.UUID, 5)
	for i := 0; i < 5; i++ {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		personIds[i] = id

		name := "name" + strconv.Itoa(i)
		surname := "surname" + strconv.Itoa(i)
		patronymic := "patronymic" + strconv.Itoa(i)
		age := 10 * (i + 1)
		gender := models.Male
		if i%2 == 0 {
			gender = models.Female
		}
		country := "RU"

		_, err := pool.Exec(ctx, queryCreatePerson,
			id,
			name,
			surname,
			patronymic,
			age,
			gender,
			country,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create person: %w", err)
		}
	}

	return personIds, nil
}

func TestGetPersonById(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := NewPersonDriver(pool)
	ctx := context.Background()

	personIds, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, personIds)

	t.Run("GetPersonById with existing id", func(t *testing.T) {
		expName := "name0"
		expSurname := "surname0"
		expPatronymic := "patronymic0"
		expCountry := "RU"
		var expAge uint32 = 10
		expGender := models.Female

		person, err := driver.GetPersonById(ctx, personIds[0])
		require.NoError(t, err)
		assert.Equal(t, expName, person.Name)
		assert.Equal(t, expSurname, person.Surname)
		assert.Equal(t, expPatronymic, person.Patronymic)
		assert.Equal(t, expAge, person.Age)
		assert.Equal(t, expGender, person.Gender)
		assert.Equal(t, expCountry, person.Country)
	})

	t.Run("GetPersonById with invalid id", func(t *testing.T) {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}

		person, err := driver.GetPersonById(ctx, id)
		require.Error(t, err)
		require.Nil(t, person)
		require.Equal(t, custom_errors.ErrPersonNotFound, err)
	})
}

func TestCreatePerson(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := NewPersonDriver(pool)
	ctx := context.Background()

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}

	person := &models.Person{
		Id:         id,
		Name:       "name",
		Surname:    "surname",
		Patronymic: "patronymic",
		Age:        10,
		Gender:     models.Male,
		Country:    "RU",
	}

	err := driver.CreatePerson(ctx, person)
	require.NoError(t, err)

	expPerson, err := driver.GetPersonById(ctx, person.Id)

	require.NoError(t, err)
	assert.Equal(t, person, expPerson)
}

func TestUpdatePerson(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := NewPersonDriver(pool)
	ctx := context.Background()

	personIds, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, personIds)

	t.Run("UpdatePerson with existing id", func(t *testing.T) {
		var age uint32 = 25
		country := "EN"
		updatePersonDto := dtos.PersonDto{
			Id:      personIds[0],
			Age:     &age,
			Country: &country,
		}

		updatedPerson, err := driver.UpdatePerson(ctx, updatePersonDto)
		require.NoError(t, err)
		assert.Equal(t, personIds[0], updatedPerson.Id)
		assert.Equal(t, age, updatedPerson.Age)
		assert.Equal(t, country, updatedPerson.Country)
	})

	t.Run("UpdatePerson without updating fields", func(t *testing.T) {
		updatePersonDto := dtos.PersonDto{
			Id: personIds[0],
		}

		updatedPerson, err := driver.UpdatePerson(ctx, updatePersonDto)
		require.Error(t, err)
		require.Nil(t, updatedPerson)
		require.Equal(t, custom_errors.ErrNoFieldsToUpdate, err)
	})
}

func TestDeletePerson(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := NewPersonDriver(pool)
	ctx := context.Background()

	personIds, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, personIds)

	err = driver.DeletePerson(ctx, personIds[0])
	require.NoError(t, err)
}

func TestGetPersons(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := NewPersonDriver(pool)
	ctx := context.Background()

	personIds, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, personIds)

	t.Run("GetPersons without filters", func(t *testing.T) {
		getPersonDto := dtos.GetPersonDto{}

		persons, err := driver.GetPersons(ctx, getPersonDto)
		require.NoError(t, err)
		require.NotEmpty(t, persons)
		require.Equal(t, len(personIds), len(persons))
	})

	t.Run("GetPersons with one filter", func(t *testing.T) {
		names := []string{"name0", "name1"}
		getPersonDto := dtos.GetPersonDto{
			Names: names,
		}

		persons, err := driver.GetPersons(ctx, getPersonDto)
		require.NoError(t, err)
		require.NotEmpty(t, persons)
		require.Equal(t, 2, len(persons))
		require.True(t, slices.Contains(names, persons[0].Name))
		require.True(t, slices.Contains(names, persons[1].Name))
	})

	t.Run("GetPersons with multiple filters", func(t *testing.T) {
		var age uint32 = 35
		getPersonDto := dtos.GetPersonDto{
			Names:     []string{"name2", "name4"},
			LowAge:    &age,
			Countries: []string{"RU", "EN"},
		}

		persons, err := driver.GetPersons(ctx, getPersonDto)
		require.NoError(t, err)
		require.NotEmpty(t, persons)
		require.Equal(t, 1, len(persons))
		require.Equal(t, "name4", persons[0].Name)
		require.Less(t, age, persons[0].Age)
		require.Equal(t, "RU", persons[0].Country)
	})
}
