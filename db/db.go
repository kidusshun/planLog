package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file" // Import the file source driver
	_ "github.com/lib/pq"                                // PostgreSQL driver
)

func NewMySQLStorage(user, password, host, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err, "is this error")
	}

	return db, err
}
