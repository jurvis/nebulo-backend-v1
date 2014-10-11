package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	DB_USERNAME = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "nebulo"
	DB_HOST     = "localhost"
)

func Dbconnect() (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		DB_USERNAME,
		DB_NAME,
		DB_PASSWORD,
		DB_HOST))

	return db, err
}
