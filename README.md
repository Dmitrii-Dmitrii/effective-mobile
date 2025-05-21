# Effective Mobile

## Описание 
Данный проект является backend-сервисом, который будет получать по API ФИО, из открытых API обогащать ответ наиболее вероятными возрастом, полом и национальностью и сохранять данные в БД. По запросу выдавать инфу о найденных людях.

## Инструкция по сборке
Для запуска сервера необходимо склонировать данный репозиторий и перейти в его директорию:
```
git clone https://github.com/Dmitrii-Dmitrii/effective-mobile.git
cd effective-mobile
```
Далее нужно перейти в директорию `/docker`. Внутри нее находится `docker-compose.yml`, который отвечает за развертывание базы данных. Его нужно запустить:
```
cd docker
docker-compose up -d
```
Далее нужно применить миграции из корневой директории:
```
goose -dir ./migrations postgres "<CONNECTION_STRING>" up
```
Переменная `<CONNECTION_STRING>` зависит от данных в `docker-compose.yml`, если ничего не менять, то нужно использовать `postgres://em_user:em_password@localhost:5432/em_database?sslmode=disable`.

После важно создать `.env` файл **в корне проекта**:
```
touch .env
```
Пример `.env` файла:
```
AGE_URL="https://api.agify.io/?name="
GENDER_URL="https://api.genderize.io/?name="
COUNTRY_URL="https://api.nationalize.io/?name="

SERVER_PORT="8080"
DB_CONNECTION_STRING="postgres://em_user:em_password@localhost:5432/em_database?sslmode=disable"
LOG_LEVEL="debug"
ENV="production"
```
Далее необходимо запустить сервер из корневой директории:
```
go run cmd/server/main.go
```

## API 
Доступны следующие методы:
- CreatePerson (`POST: /persons`) - создание нового человека;
- UpdatePerson (`PUT: /persons`) - обновление человека;
- DeletePerson (`DELETE: /persons/:id`) - удаление человека;
- GetPersons (`GET: /persons`) - получение всех людей с фильтрацией;
- GetPersonById (`GET: /persons/:id`) - получение человека по его id.

Более подробную информацию об API можно получить, перейдя по `/swagger/index.html`.