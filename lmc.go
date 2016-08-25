package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	u "os/user"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	usr, err := u.Current()
	if err != nil {
		fmt.Printf("Cannot get current OS user details: %s\n", err.Error())
		return
	}
	dir := usr.HomeDir
	config := &Config{
		Dir: dir + string(filepath.Separator) + ".lmc",
		AuthTokenFilename: "auth.token",
		CredentialsFilename: "credentials",
		APIHost: "http://localhost:8080/api/",
	}
	userCredentials, err := setup(config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	auth := Auth{}
	auth.Config = config
	auth.UserCredentials = userCredentials
	// colors
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	blue := color.New(color.FgBlue)

	reader := bufio.NewReader(os.Stdin)
	for {
		blue.Print("cmd> ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		args := strings.FieldsFunc(cmd, func(c rune) bool {
			return unicode.IsSpace(c)
		})
		if len(args) == 0 {
			continue
		}
		// Identify command by first argument
		switch strings.TrimSpace(args[0]) {
		case "ua":
			newUser, err := addUser(&auth)
			if err == nil {
				blue.Printf("User created %v\n", newUser)
			} else {
				red.Printf("%v\n", err)
			}
		case "ia":
			blue.Println("We are going to add link")
			link, err := ParseLink(args[1:])
			if err != nil {
				red.Printf("%v\n", err)
			} else {
				blue.Printf("Parsed link: %v\n", link)
			}
		case "auth":
			_, err := auth.GetToken(true)
			if err != nil {
				red.Printf("%v\n", err)
			} else {
				green.Printf("Authorised OK. User ID: %s\n", auth.UserAuth.User.ID)
			}
		case "credentials":
			_, _, err = readAndSaveUserCredentials(config.Dir + string(filepath.Separator) + config.CredentialsFilename)
			if err == nil {
				blue.Println("Credentials saved")
			} else {
				red.Printf("%v\n", err)
			}
		case "exit":
			blue.Println("Bye!")
			return
		default:
			if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
				blue.Printf("We are going to add link %s\n", args[0])
				link, err := ParseLink(args)
				if err != nil {
					red.Printf("%v\n", err)
				} else {
					switch link.(type) {
					case *Link:
						link, err := addLink(&auth, link.(*Link))
						if err == nil {
							blue.Printf("Item created %v\n", link)
						} else {
							red.Printf("%v\n", err)
						}
					default:
						red.Println("Unknown item type: %v", link)
					}
				}
			}
		}
	}
}

func setup(config *Config) (*UserCredentials, error) {
	// Check if folder exists
	if _, err := os.Stat(config.Dir); os.IsNotExist(err) {
		err = os.MkdirAll(config.Dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("Configuration folder %s could not be created: %s", config.Dir, err.Error())
		}
	}
	// Create auth token file
	authTokenFilename := config.Dir + string(filepath.Separator) + config.AuthTokenFilename
	credentialsFilename := config.Dir + string(filepath.Separator) + config.CredentialsFilename
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
