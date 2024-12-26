package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"library/handler"

	"github.com/gorilla/schema"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(1, 3)

func main() {
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v\n", err)
	}
	defer logFile.Close()

	logger := logrus.New()
	logger.SetOutput(logFile)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting application...")

	var createTable = `
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

	logger.WithField("time", time.Now()).Info("Connecting to the database...")
	db, err := sqlx.Connect("postgres", "user=postgres password=gakon dbname=library_system sslmode=disable")
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to the database")
	}
	logger.Info("Successfully connected to the database")

	logger.Info("Creating tables if they don't exist...")
	_, err = db.Exec(createTable)
	if err != nil {
		logger.WithError(err).Fatal("Error creating tables")
	}
	logger.Info("Tables created or already exist")

	logger.Info("Initializing form decoder...")
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	logger.Info("Form decoder initialized")

	logger.Info("Setting up session store...")
	store := sessions.NewCookieStore([]byte("jsowjpw38eowj4ur82jmaole0uehqpl"))
	logger.Info("Session store set up")

	logger.Info("Initializing handler...")
	r := handler.New(db, decoder, store, logger) // Передаём логгер в обработчики

	r.Use(rateLimiterMiddleware(limiter, logger))

	serverAddress := "127.0.0.1:3000"
	logger.WithField("address", serverAddress).Info("Server starting...")
	if err := http.ListenAndServe(serverAddress, r); err != nil {
		logger.WithError(err).Fatal("Server failed to start")
	}

}

func rateLimiterMiddleware(limiter *rate.Limiter, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				logger.WithField("remote_ip", r.RemoteAddr).Warn("Rate limit exceeded for request")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
