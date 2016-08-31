package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Auth represent authentication info.
type Auth struct {
	Config          *Config
	UserCredentials *UserCredentials
	Token           string
}

// GetToken read saved session token, check expiration date and request new token if needed.
func (a *Auth) GetToken() (string, error) {
	var err error
	if a.Token == "" {
		a.Token, err = a.readToken()
		if err != nil {
			return "", fmt.Errorf("Getting token failed: %s", err.Error())
		}
	}
	return a.Token, nil
}

// Authenticate uses API object to request new token.
func (a *Auth) Authenticate() (string, error) {
	var err error
	api := API{a.Config.APIHost}
	a.Token, err = api.Auth(a.UserCredentials.Username, a.UserCredentials.Password)
	if err != nil {
		return "", fmt.Errorf("Failed to authenticate: %s", err.Error())
	}
	err = writeToken(a.Config.Dir+string(filepath.Separator)+a.Config.AuthTokenFilename, a.Token)
	if err != nil {
		return "", fmt.Errorf("Writing token failed: %s", err.Error())
	}

	return a.Token, nil
}

// readToken read token info from special file.
func (a *Auth) readToken() (string, error) {
	at, err := ioutil.ReadFile(a.Config.Dir + string(filepath.Separator) + a.Config.AuthTokenFilename)
	if err != nil {
		return "", fmt.Errorf("Reading token from file failed: %s", err.Error())
	}
	return strings.TrimSpace(string(at)), nil
}

// writeToken saves auth token to file.
func writeToken(filename, token string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}
	defer f.Close()
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}
	_, err = f.WriteString(token)
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}

	return nil
}
