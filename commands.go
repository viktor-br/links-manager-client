package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func addUser(auth *Auth) (*User, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please provide new user details")
	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Could not read new user username: %s", err.Error())
	}
	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("Could not read new user password: %s", err.Error())
	}
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	newUser := &User{Username: username, Password: password}

	api := API{auth.Config.APIHost}
	err = authenticateWrapper(auth, func(token string) error {
		err = api.UserAdd(token, newUser)

		return err
	})

	return newUser, err
}

func addLink(auth *Auth, link *Link) (*Link, error) {
	api := API{auth.Config.APIHost}
	err := authenticateWrapper(auth, func(token string) error {
		_, err := api.LinkAdd(token, link)

		return err
	})

	return nil, err
}

func checkConnection(auth *Auth) (bool, error) {
	api := API{auth.Config.APIHost}
	return api.Ping()
}

// authenticateWrapper add feature of re-authentication in case of HTTP error 401
func authenticateWrapper(auth *Auth, p func(string) error) error {
	token, err := auth.GetToken()
	if err != nil {
		return err
	}
	err = p(token)
	if err != nil {
		if e, ok := err.(*APIError); ok {
			if e.code == ErrUnauthorized {
				token, err := auth.Authenticate()
				if err != nil {
					return err
				}
				return p(token)
			}
		}
		return err
	}
	return nil
}
