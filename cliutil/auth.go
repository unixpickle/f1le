package cliutil

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

// PrintEnvUsage writes instructions on setting up authentication
// via environment variables to stderr.
func PrintEnvUsage() {
	fmt.Fprintln(os.Stderr, "Set F1LE_ROOT and F1LE_PASS environment variables.")
	fmt.Fprintln(os.Stderr, "F1LE_ROOT is a URL, like 'http://localhost:1337'.")
}

// AuthEnvClient creates a client that is authenticated
// based on configuration environment variables.
//
// Also returns the server base URL.
func AuthEnvClient() (*http.Client, *url.URL, error) {
	baseURL, password, err := ReadEnv()
	if err != nil {
		return nil, nil, err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}
	client := &http.Client{Jar: jar}
	if err := Authenticate(client, *baseURL, password); err != nil {
		return nil, nil, err
	}
	return client, baseURL, nil
}

// ReadEnv reads authentication environment variables, or returns
// an error if the variables were invalid.
func ReadEnv() (baseURL *url.URL, password string, err error) {
	baseStr := os.Getenv("F1LE_ROOT")
	password = os.Getenv("F1LE_PASS")
	if baseStr == "" || password == "" {
		err = errors.New("must specify F1LE_ROOT and F1LE_PASS environment variables")
		return
	}
	baseURL, err = url.Parse(baseStr)
	if err != nil {
		err = fmt.Errorf("invalid F1LE_ROOT: %w", err)
	}
	return
}

// Authenticate configures the client to access a server
// with the base URL and password, returning an error if
// the login fails.
func Authenticate(c *http.Client, u url.URL, password string) error {
	u.Path = "/login"
	vals := url.Values{}
	vals.Set("password", password)
	resp, err := c.PostForm(u.String(), vals)
	if err != nil {
		return fmt.Errorf("failed to send authentication request: %w", err)
	}
	resp.Body.Close()
	if resp.Request.URL.Path == "/login" {
		return errors.New("login failed")
	}
	return nil
}
