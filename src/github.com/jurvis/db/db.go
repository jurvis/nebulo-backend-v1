package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func Dbconnect() (*sql.DB, error) {
	cfg := config.Config()

	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Dbname,
		cfg.Database.Password,
		cfg.Database.Host))

	return db, err
}
