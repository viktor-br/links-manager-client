package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/satori/go.uuid"
	s "github.com/viktor-br/jobs-scheduler"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	u "os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
		Dir:                     dir + string(filepath.Separator) + ".lmc",
		AuthTokenFilename:       "auth.token",
		CredentialsFilename:     "credentials",
		APIHost:                 "http://localhost:8080/api/",
		LogFilename:             "links-manager-client.log",
		UncompletedJobsFilename: "uncompleted-jobs.json",
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

	var buf bytes.Buffer
	// Init logger with output to file
	f, err := os.OpenFile(config.LogPath(), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Failed to open file: %v", err)
		return
	}
	err = f.Truncate(0)
	if err != nil {
		fmt.Printf("Failed to clear log file: %v", err)
		return
	}
	defer f.Close()

	logger := log.New(&buf, "", log.Ldate|log.Ltime|log.LUTC)
	logger.SetOutput(f)

	unprocessedJobs := []Job{}
	jobs := make(chan Job, 10)
	noConnection := make(chan bool)

	// Create scheduler with simple processor, which sleeps 3 seconds to emulate it's doing something.
	// TODO check if it worth it to use closure to pass authentication
	scheduler := s.NewJobsScheduler(func(job s.Job) s.JobResult {
		jobResult := JobResult{}
		jobResult.job = job.(Job)
		switch job.(type) {
		case Job:
			_, err := addLink(&auth, job.(Job).Link)
			if err != nil {
				jobResult.lastError = err
			}
		default:
			jobResult.lastError = fmt.Errorf("Unknow job type #%s", job.GetID())
		}
		return jobResult
	})
	// Set up options
	scheduler.Option(s.MaxTries(3), s.ProcessorsNum(2))
	scheduler.AddLogger(func(msg string) {
		logger.Println(msg)
	})
	// Add function which process results flow
	scheduler.AddResultOutput(func(res s.JobResult) {
		switch res.(type) {
		case JobResult:
			jobResult := res.(JobResult)
			if !jobResult.IsDone() {
				unprocessedJobs = append(unprocessedJobs, jobResult.job)
				logger.Printf("job #%s failed: %s\n", res.GetJobID(), res.(JobResult).lastError.Error())
				// Send signal, connection failed, so we need to stop send requests and wait reestablishing connection.
				if jobResult.ConnectionFailed() {
					noConnection <- true
				}
			} else {
				logger.Printf("job #%s successed\n", res.GetJobID())
			}
		default:
			logger.Println("unknown result type")
		}
	})
	scheduler.Run()

	// Read previously saved uncompleted jobs from file
	savedJobs := []Job{}
	raw, err := ioutil.ReadFile(config.UncompletedJobsPath())
	if err != nil {
		fmt.Println("Cannot read uncompleted jobs file")
	}
	json.Unmarshal(raw, &savedJobs)

	// Schedule uncompleted jobs
	for _, v := range savedJobs {
		err = scheduler.Add(v)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	reader := bufio.NewReader(os.Stdin)
	// Buffer = 1 b/c no need to block the goroutine.
	signalsDone := make(chan bool, 1)
	// Wait for Ctrl+C close signal
	go func(scheduler *s.JobsScheduler, signalsDone chan bool) {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals
		fmt.Println()
		scheduler.Shutdown()
		signalsDone <- true
	}(scheduler, signalsDone)

	// Run separate goroutine, which accepts a job and forward it either to scheduler or to local storage.
	go schedule(&auth, scheduler, noConnection, jobs)

	// Command line goroutine, read and run command
	go func(scheduler *s.JobsScheduler) {
	Exit:
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
			case "auth":
				_, err := auth.Authenticate()
				if err != nil {
					red.Printf("%v\n", err)
				} else {
					green.Println("Authorised OK")
				}
			case "credentials":
				_, _, err = readAndSaveUserCredentials(config.CredentialsPath())
				if err == nil {
					blue.Println("Credentials saved")
				} else {
					red.Printf("%v\n", err)
				}
			case "exit":
				scheduler.Shutdown()
				signalsDone <- true
				break Exit
			case "ping":
				if checkConnection(&auth) {
					green.Println("Ok: server is available")
				} else {
					red.Println("Failed: server is not available")
				}
			default:
				// If command starts with url, user wants to add link
				if strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://") {
					link, err := ParseLink(args)
					if err != nil {
						red.Printf("%v\n", err)
					} else {
						switch link.(type) {
						case *Link:
							jobs <- Job{ID: uuid.NewV4().String(), Link: link.(*Link)}
						default:
							red.Println("Unknown item type: %v", link)
						}
					}
				}
			}
		}
	}(scheduler)

	<-signalsDone

	scheduler.Wait()

	// Read from scheduler jobs, which were not processed and save to file
	for _, j := range scheduler.GetUncompletedJobs() {
		unprocessedJobs = append(unprocessedJobs, j.(Job))
	}
	b, err := json.Marshal(unprocessedJobs)
	if err == nil {
		ioutil.WriteFile(config.UncompletedJobsPath(), b, 0644)
	}
}

func schedule(auth *Auth, scheduler *s.JobsScheduler, noConnection chan bool, jobs chan Job) {
	successChan := make(chan bool)
	connectionFailed := false
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	for {
		select {
		case <-noConnection:
			// Ignore a channel failure while processing a previous one still in progress.
			if !connectionFailed {
				connectionFailed = true
				// Run goroutine to periodically check if the server is available.
				go waitForServerAvailable(auth, successChan)
			}
		case <-successChan:
			connectionFailed = true
			// TODO Now we can read jobs from local storage and resend.
		case job := <-jobs:
			if connectionFailed {
				// TODO Save to the local storage
			} else {
				// Send the job to the scheduler
				err := scheduler.Add(job)
				if err == nil {
					green.Printf("AddLink for %s scheduled\n", job.Link)
				} else {
					red.Printf("Parsed link: %v\n", err)
				}
			}
		}
	}
}

// waitForServerAvailable requests a server endpoint and exit in case the server is available.
func waitForServerAvailable(auth *Auth, success chan bool) {
	timeout := 15 * time.Second
	defer func() {
		success <- true
	}()
	for {
		if checkConnection(auth) {
			break
		} else {
			time.Sleep(timeout)
			timeout = timeout * 2
		}
	}
}

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
