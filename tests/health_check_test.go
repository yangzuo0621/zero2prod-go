package tests

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/yangzuo0621/zero2prod-go/app"
)

// spin up an instance of our application
// and returns its address (i.e. http://localhost:XXXX)
// https://www.lifewire.com/port-0-in-tcp-and-udp-818145
// Port 0 is special-cased at the OS level: trying to bind port 0 will trigger an OS scan for an available
// port which will then be bound to the application.
func spawnApp() (*TestApp, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to bind random port")
	}

	configuration, err := app.GetConfiguration("../")
	if err != nil {
		return nil, fmt.Errorf("Failed to read configuration")
	}

	database, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("Failed to create database name: %w", err)
	}
	configuration.Database.DatabaseName = database.String()

	db, err := configureDatabase(&configuration.Database)
	if err != nil {
		return nil, fmt.Errorf("configureDatabase: %w", err)
	}

	go func() {
		err := app.Run(listener, db)
		if err != nil {
			fmt.Println("failed to start app:", err)
		}
	}()

	return &TestApp{
		address: listener.Addr().String(),
		db:      db,
	}, nil
}

type TestApp struct {
	address string
	db      *sql.DB
}

func configureDatabase(config *app.DatabaseSettings) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.ConnectionStringWithoutDB())
	if err != nil {
		return nil, fmt.Errorf("Failed to open to Postgres: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to postgres: %w", err)
	}
	_, err = db.Exec("CREATE DATABASE \"" + config.DatabaseName + "\"")
	if err != nil {
		return nil, fmt.Errorf("Failed to create database: %w", err)
	}

	db, err = sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("Failed to open to Postgres: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get instance driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file:///home/yangzuo/workspace/projects_go/zero2prod-go/migrations/", config.DatabaseName, driver)
	if err != nil {
		return nil, fmt.Errorf("Failed to get migration instance: %w", err)
	}

	err = m.Up()
	if err != nil {
		return nil, fmt.Errorf("Failed to migrate: %w", err)
	}

	return db, nil
}

func TestHealthCheckWorks(t *testing.T) {
	// Arrange
	testApp, err := spawnApp()
	assert.NoError(t, err, "Failed to start app")
	client := http.Client{}

	// Act
	// response, err := http.Get(fmt.Sprintf("http://%s/health_check", address))
	// assert.NoError(t, err)

	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://%s/health_check", testApp.address),
		nil,
	)
	assert.NoError(t, err, "Failed to create request.")

	response, err := client.Do(request)
	assert.NoError(t, err, "Failed to execute request.")

	// Assert
	assert.Equal(t, response.StatusCode, http.StatusOK)
	assert.EqualValues(t, response.ContentLength, 0)
}

func TestSubscribeReturn200ForValidFormData(t *testing.T) {
	// Arrange
	testApp, err := spawnApp()
	assert.NoError(t, err, "Failed to start app")

	client := http.Client{}

	// Act
	body := "name=le%20guin&email=ursula_le_guin%40gmail.com"
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s/subscriptions", testApp.address),
		strings.NewReader(body),
	)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err, "Failed to create request.")

	response, err := client.Do(request)
	assert.NoError(t, err, "Failed to execute request.")

	// Assert
	assert.Equal(t, response.StatusCode, http.StatusOK)

	var (
		email string
		name  string
	)
	err = testApp.db.QueryRow("SELECT email, name FROM subscriptions").Scan(&email, &name)
	assert.NoError(t, err, "Failed to fetch saved subscription.")
	assert.Equal(t, email, "ursula_le_guin@gmail.com")
	assert.Equal(t, name, "le guin")
}

func TestSubscribeReturns400WhenDataIsMissing(t *testing.T) {
	// Arrange
	testApp, err := spawnApp()
	assert.NoError(t, err)
	client := http.Client{}
	testCases := []struct { // table-driven test https://github.com/golang/go/wiki/TableDrivenTests
		invalidBody  string
		errorMessage string
	}{
		{"name=le%20guin", "missing the email"},
		{"email=ursula_le_guin%40gmail.com", "missing the name"},
		{"", "missing both name and email"},
	}

	for _, testCase := range testCases {
		// Act
		request, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("http://%s/subscriptions", testApp.address),
			strings.NewReader(testCase.invalidBody),
		)
		assert.NoError(t, err, "Failed to create request.")
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		response, err := client.Do(request)
		assert.NoError(t, err, "Failed to execute request.")

		// Assert
		assert.Equalf(t, response.StatusCode, http.StatusBadRequest,
			"The API did not fail with 400 Bad Request when the payload was %s.",
			testCase.errorMessage,
		)
	}
}
