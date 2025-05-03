package db

import (
	"fmt"
	"log"
	"net/url"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
)

func NewDriver(user string, pass string, host string, port int, db string) *sql.Driver {
	dsn := url.URL{
		Scheme: dialect.Postgres,
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   db,
		RawQuery: url.Values{
			"sslmode": []string{"disable"},
		}.Encode(),
	}

	// Open connection to database
	driver, err := sql.Open(dialect.Postgres, dsn.String())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return driver
}
