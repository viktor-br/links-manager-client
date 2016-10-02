package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// ErrUnauthorized API error code
	ErrUnauthorized = 401
)

// API struct to manager requests to remote
type API struct {
	Host string
}

// APIError is a custom error, to handle exceptions outside API calls.
type APIError struct {
	code int
}

func (err *APIError) Error() string {
	return fmt.Sprintf("%d", err.code)
}

// Auth sends an authentication request to remote and return token.
func (a *API) Auth(username string, password string) (string, error) {
	url := a.Host + "user/login"
	user := User{Username: username, Password: password}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(user)
	if err != nil {
		return "", fmt.Errorf("Unable to encode user %s: %s", username, err.Error())
	}
	res, err := http.Post(url, "application/json", b)
	if err != nil {
		if res != nil {
			res.Body.Close()
			return "", fmt.Errorf("Auth POST request failed with status %s: %s", res.Status, err.Error())
		}
		return "", fmt.Errorf("Auth POST request failed: %s", err.Error())
	}
	if res.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Your user %s failed to authenticate with code %d", username, res.StatusCode)
	}
	if xAuthToken := res.Header.Get("X-AUTH-TOKEN"); xAuthToken != "" {
		return xAuthToken, nil
	}

	return "", fmt.Errorf("Token is empty for username: %s", res.StatusCode)
}

// UserAdd sends a request to remote to create new user.
func (a *API) UserAdd(token string, newUser *User) error {
	url := a.Host + "user"
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(newUser)
	if err != nil {
		return fmt.Errorf("Unable to encode new user %s: %s", newUser.Username, err.Error())
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, b)
	if err != nil {
		return fmt.Errorf("Creating UserAdd request failed for user %s: %s", newUser.Username, err.Error())
	}
	if req != nil {
		defer func() {
			req.Close = true
		}()
	}
	req.Header.Add("X-AUTH-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		if res != nil {
			res.Body.Close()
			return fmt.Errorf("UserAdd failed with status %s: %s", res.Status, err.Error())
		}
		return fmt.Errorf("API::UserAdd failed: %s", res.Status, err.Error())
	}

	return nil
}

// LinkAdd sends a request to create new link item.
func (a *API) LinkAdd(token string, link *Link) (*Link, error) {
	url := a.Host + "item/link"
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(link)
	if err != nil {
		return nil, fmt.Errorf("Unable to encode new link %s: %s", link.URL, err.Error())
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, b)
	if err != nil {
		return nil, fmt.Errorf("Creating itemAdd request failed for item %s: %s", link.URL, err.Error())
	}
	if req != nil {
		defer func() {
			req.Close = true
		}()
	}
	req.Header.Add("X-AUTH-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		if res != nil {
			res.Body.Close()
			return nil, fmt.Errorf("ItemAdd failed with status %s: %s", res.Status, err.Error())
		}
		return nil, fmt.Errorf("API::ItemAdd failed: %s", err.Error())
	}
	if res.StatusCode == http.StatusUnauthorized {
		return nil, &APIError{http.StatusUnauthorized}
	}
	// TODO If response is 200, parse response body into new item object
	bt := []byte{}
	res.Body.Read(bt)

	return nil, nil
}

// Ping sends request to special endpoint to check if server is available.
func (a *API) Ping() bool {
	url := a.Host + "ping"
	resp, _ := http.Get(url) // if error occurred, ping failed, no need to know why
	if resp != nil {
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}

	return false
}
