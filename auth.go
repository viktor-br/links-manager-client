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
	UserID  string
	Config  *Config
	User    *User
	token   string
	expires int64 `json:"expires"`
}

// ExtractExpiresAndUserID parse token to get expires and user ID fields
func ExtractExpiresAndUserID(token string) (string, int64, error) {
	ua := &UserAuth{}
	parts := strings.Split(token, ".")
	if len(parts) > 0 && parts[0] != "" {
		u, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return "", 0, fmt.Errorf("Token %s decode failed: %s", token, err.Error())
		}

		err = json.NewDecoder(bytes.NewReader(u)).Decode(ua)
		if err != nil {
			return "", 0, fmt.Errorf("Token %s json decode failed: %s", token, err.Error())
		}
		if ua.User == nil {
			return "", 0, fmt.Errorf("Cannot parse user data: %s", token, err.Error())
		}

		return ua.User.ID, ua.Expires, nil
	}
	return "", 0, nil
}

// GetToken read saved session token, check expiration date and request new token if needed.
func (a *Auth) GetToken(forceAPIRequest bool) (string, error) {
	var err error
	if a.expires == 0 {
		err = a.readToken()
		if err != nil {
			return "", fmt.Errorf("Getting token failed: %s", err.Error())
		}
	}
	if a.expires <= time.Now().Unix() || forceAPIRequest {
		a.token = ""
		api := API{a.Config}
		a.token, err = api.Auth(a.User.Username, a.User.Password)
		if err != nil {
			return "", fmt.Errorf("Failed to authenticate: %s", err.Error())
		}
		a.UserID, a.expires, err = ExtractExpiresAndUserID(a.token)
		if err != nil {
			return "", fmt.Errorf("Extracting token expires failed: %s", err.Error())
		}

		err = a.writeToken()
		if err != nil {
			return "", fmt.Errorf("Writing token failed: %s", err.Error())
		}
	}

	return a.token, nil
}

// readToken read token info from special file.
func (a *Auth) readToken() error {
	at, err := ioutil.ReadFile(a.Config.Dir + string(filepath.Separator) + a.Config.AuthTokenFilename)
	if err != nil {
		return fmt.Errorf("Reading token from file failed: %s", err.Error())
	}
	str := strings.TrimSpace(string(at))
	a.token = str
	a.UserID, a.expires, err = ExtractExpiresAndUserID(str)
	if err != nil {
		return fmt.Errorf("Reading token from file failed: %s", err.Error())
	}

	return nil
}

// writeToken saves auth token to file.
func (a *Auth) writeToken() error {
	f, err := os.Create(a.Config.Dir + string(filepath.Separator) + a.Config.AuthTokenFilename)
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}
	defer f.Close()
	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}
	_, err = f.WriteString(a.token)
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("Writing token to file failed: %s", err.Error())
	}

	return nil
}
