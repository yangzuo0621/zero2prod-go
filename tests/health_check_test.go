package tests

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yangzuo0621/zero2prod-go/app"
)

// spin up an instance of our application
// and returns its address (i.e. http://localhost:XXXX)
// https://www.lifewire.com/port-0-in-tcp-and-udp-818145
// Port 0 is special-cased at the OS level: trying to bind port 0 will trigger an OS scan for an available
// port which will then be bound to the application.
func spawnApp() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to bind random port")
	}

	go func() {
		err := app.Run(listener)
		if err != nil {
			fmt.Println("failed to start app:", err)
		}
	}()

	return listener.Addr().String(), nil
}

func TestHealthCheckWorks(t *testing.T) {
	// Arrange
	address, err := spawnApp()
	assert.NoError(t, err, "Failed to start app")
	client := http.Client{}

	// Act
	// response, err := http.Get(fmt.Sprintf("http://%s/health_check", address))
	// assert.NoError(t, err)

	request, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://%s/health_check", address),
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
	address, err := spawnApp()
	assert.NoError(t, err, "Failed to start app")
	client := http.Client{}

	// Act
	body := "name=le%20guin&email=ursula_le_guin%40gmail.com"
	request, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s/subscriptions", address),
		strings.NewReader(body),
	)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err, "Failed to create request.")

	response, err := client.Do(request)
	assert.NoError(t, err, "Failed to execute request.")

	// Assert
	assert.Equal(t, response.StatusCode, http.StatusOK)
}

func TestSubscribeReturns400WhenDataIsMissing(t *testing.T) {
	// Arrange
	address, err := spawnApp()
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
			fmt.Sprintf("http://%s/subscriptions", address),
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
