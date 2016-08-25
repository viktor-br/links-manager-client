package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func addUser(config *Config, user *User) (*User, error) {
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

	auth := Auth{}
	auth.Config = config
	auth.User = user
	token, err := auth.GetToken(false)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}
	api := API{config}
	err = api.UserAdd(token, newUser)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}

	return newUser, nil
}

func addLink(config *Config, user *User, link *Link) (*Link, error) {
	auth := Auth{}
	auth.Config = config
	auth.User = user
	token, err := auth.GetToken(false)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}
	api := API{config}
	//link.UserId = auth.UserId
	_, err = api.LinkAdd(token, link)
	if err != nil {
		return nil, fmt.Errorf("%s\n", err.Error())
	}

	return nil, nil
}
