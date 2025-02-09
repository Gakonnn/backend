package handler

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"log"
	_ "os"
	"testing"
	_ "time"
)

var db *sqlx.DB

func setupTestDatabase() {
	var err error
	db, err = sqlx.Connect("postgres", "user=postgres password=gakon dbname=library_system sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id serial,
		name text,
		status boolean,
		primary key (id)
	);
	CREATE TABLE IF NOT EXISTS books (
		id serial,
		category_id integer,
		book_name text,
		author_name text,
		details text,
		image text,
		status boolean,
		primary Key (id)
	);
	CREATE TABLE IF NOT EXISTS bookings (
		id serial,
		user_id integer,
		book_id integer,
		start_time timestamp,
		end_time timestamp,
		primary Key (id)
	);
	CREATE TABLE IF NOT EXISTS users (
		id serial,
		first_name text,
		last_name text,
		email text,
		password text,
		is_verified boolean,
		primary Key (id)
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("Error creating tables: %v\n", err)
	}
}

func TestDatabaseConnection(t *testing.T) {
	setupTestDatabase()

	if db.Ping() != nil {
		t.Fatal("Failed to connect to the database")
	}

	var tableCount int
	err := db.Get(&tableCount, "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		t.Fatalf("Error fetching tables count: %v\n", err)
	}

	assert.Greater(t, tableCount, 0, "Expected at least one table to be created")

	_, err = db.Exec(`INSERT INTO categories (name, status) VALUES ($1, $2)`, "Science", true)
	if err != nil {
		t.Fatalf("Error inserting category: %v\n", err)
	}

	var categoryName string
	err = db.Get(&categoryName, "SELECT name FROM categories WHERE name = $1", "Science")
	if err != nil {
		t.Fatalf("Error retrieving category: %v\n", err)
	}

	assert.Equal(t, "Science", categoryName, "Category name should be 'Science'")
}

func TestRateLimiter(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	limiter := rate.NewLimiter(1, 3)

	if !limiter.Allow() {
		t.Fatal("Rate limit exceeded")
	}

	assert.True(t, limiter.Allow(), "Rate limit should allow the request")
}
