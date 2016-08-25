package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Auth represent authentication info.
type Auth struct {
	Config          *Config
	UserCredentials *UserCredentials
	UserAuth        *UserAuth
	Token           string
}

// ParseToken parse token to get expires and user ID fields
func ParseToken(token string) (*UserAuth, error) {
	ua := &UserAuth{}
	parts := strings.Split(token, ".")
	if len(parts) > 0 && parts[0] != "" {
		u, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, fmt.Errorf("Token %s decode failed: %s", token, err.Error())
		}

		err = json.NewDecoder(bytes.NewReader(u)).Decode(ua)
		if err != nil {
			return nil, fmt.Errorf("Token %s json decode failed: %s", token, err.Error())
		}
		if ua.User == nil {
			return nil, fmt.Errorf("Cannot parse user data: %s", token, err.Error())
		}

		return ua, nil
	}
	return nil, fmt.Errorf("Malformatted token: %s", token)
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

// GetToken read saved session token, check expiration date and request new token if needed.
func (a *Auth) GetToken(forceAPIRequest bool) (string, error) {
	var err error
	if a.UserAuth != nil && a.UserAuth.Expires == 0 && !forceAPIRequest {
		a.Token, err = a.readToken()
		if err != nil {
			return "", fmt.Errorf("Getting token failed: %s", err.Error())
		}
		a.UserAuth, err = ParseToken(a.Token)
		if err != nil {
			return "", fmt.Errorf("Reading token from file failed: %s", err.Error())
		}
	}
	if a.UserAuth == nil || a.UserAuth.Expires <= time.Now().Unix() || forceAPIRequest {
		a.Token = ""
		api := API{a.Config.APIHost}
		a.Token, err = api.Auth(a.UserCredentials.Username, a.UserCredentials.Password)
		if err != nil {
			return "", fmt.Errorf("Failed to authenticate: %s", err.Error())
		}
		a.UserAuth, err = ParseToken(a.Token)
		if err != nil {
			return "", fmt.Errorf("Extracting token expires failed: %s", err.Error())
		}

		err = writeToken(a.Config.Dir+string(filepath.Separator)+a.Config.AuthTokenFilename, a.Token)
		if err != nil {
			return "", fmt.Errorf("Writing token failed: %s", err.Error())
		}
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
