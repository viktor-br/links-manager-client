package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"strings"
)


// setup creates required files and read data from the previously saved files (log, credentials and authentication info).
func setup(config *Config) (*UserCredentials, error) {
	// Check if folder exists
	if _, err := os.Stat(config.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(config.Dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Configuration folder %s could not be created: %s", config.Dir, err.Error())
		}
	}
	// Create auth token file
	authTokenFilename := config.AuthTokenPath()
	credentialsFilename := config.CredentialsPath()
	if _, err := os.Stat(authTokenFilename); os.IsNotExist(err) {
		err = ioutil.WriteFile(authTokenFilename, []byte{}, 0600)
		if err != nil {
			return nil, fmt.Errorf("Could not create token file %s: %s", authTokenFilename, err.Error())
		}
	}
	if _, err := os.Stat(credentialsFilename); os.IsNotExist(err) {
		username, password, err := readAndSaveUserCredentials(credentialsFilename)
		if err != nil {
			return nil, fmt.Errorf("Cannot read and save credentials: %s", err.Error())
		}

		return &UserCredentials{Username: username, Password: password}, nil
	}
	c, err := ioutil.ReadFile(credentialsFilename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read data from the credentials file: %s", err.Error())
	}
	parts := strings.Split(string(c), ":")
	if len(parts) == 2 {
		return &UserCredentials{Username: parts[0], Password: parts[1]}, nil
	}
	// read credentials and save
	username, password, err := readAndSaveUserCredentials(credentialsFilename)
	if err != nil {
		return nil, fmt.Errorf("Cannot read and save credentials: %s", err.Error())
	}
	return &UserCredentials{Username: username, Password: password}, nil
}

// readAndSaveUserCredentials requests credentials from user and save to the file.
func readAndSaveUserCredentials(credentialsFilename string) (username, password string, err error) {
	blue := color.New(color.FgBlue)
	// Read credentials and save
	reader := bufio.NewReader(os.Stdin)
	blue.Println("Please provide your credentials")
	fmt.Print("Enter username: ")
	username, err = reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("Failed to read username from console: %s", err.Error())
	}
	fmt.Print("Enter password: ")
	password, err = reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("Failed to read password from console: %s", err.Error())
	}
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)

	f, err := os.Create(credentialsFilename)
	if err != nil {
		return "", "", fmt.Errorf("Failed to create credentials file: %s", err.Error())
	}
	defer f.Close()
	err = f.Chmod(0600)
	if err != nil {
		return "", "", fmt.Errorf("Failed to set permissions to the credentials file: %s", err.Error())
	}
	err = f.Truncate(0)
	if err != nil {
		return "", "", fmt.Errorf("Failed to clear credentials file: %s", err.Error())
	}
	_, err = f.WriteString(strings.Join([]string{strings.TrimSpace(username), strings.TrimSpace(password)}, ":"))
	if err != nil {
		return "", "", fmt.Errorf("Failed to write data to the credentials file: %s", err.Error())
	}

	return
}

