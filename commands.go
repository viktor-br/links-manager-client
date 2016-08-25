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

	token, err := auth.GetToken(false)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}
	api := API{auth.Config.APIHost}
	err = api.UserAdd(token, newUser)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}

	return newUser, nil
}

func addLink(auth *Auth, link *Link) (*Link, error) {
	token, err := auth.GetToken(false)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}
	api := API{auth.Config.APIHost}
	_, err = api.LinkAdd(token, link)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}

	return nil, nil
}
