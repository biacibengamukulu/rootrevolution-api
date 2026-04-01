package cassandra

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

func RunMigrations(session *gocql.Session) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS products (
			org             text,
			id              int,
			app_name        text,
			name            text,
			description     text,
			category        text,
			image           text,
			price           double,
			original_price  double,
			discount_pct    double,
			stock_qty       int,
			is_new          boolean,
			is_best_seller  boolean,
			is_on_sale      boolean,
			created_at      timestamp,
			updated_at      timestamp,
			created_by      text,
			updated_by      text,
			status          text,
			PRIMARY KEY (org, id)
		)`,

		`CREATE TABLE IF NOT EXISTS users (
			id         uuid,
			name       text,
			surname    text,
			email      text,
			password   text,
			role       text,
			status     text,
			created_at timestamp,
			PRIMARY KEY (email)
		)`,

		`CREATE TABLE IF NOT EXISTS users_by_id (
			id         uuid,
			email      text,
			PRIMARY KEY (id)
		)`,

		`CREATE TABLE IF NOT EXISTS pending_updates (
			auth_token   uuid,
			product_id   int,
			org          text,
			update_data  text,
			requested_by text,
			created_at   timestamp,
			expires_at   timestamp,
			status       text,
			action       text,
			PRIMARY KEY (auth_token)
		)`,
	}

	for _, q := range migrations {
		if err := session.Query(q).Exec(); err != nil {
			return fmt.Errorf("migration failed: %w\nQuery: %s", err, q)
		}
		log.Println("Migration applied successfully")
	}

	return nil
}
