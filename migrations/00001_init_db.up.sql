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