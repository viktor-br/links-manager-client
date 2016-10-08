package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/satori/go.uuid"
	s "github.com/viktor-br/jobs-scheduler"
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
		Dir:                 dir + string(filepath.Separator) + ".lmc",
		AuthTokenFilename:   "auth.token",
		CredentialsFilename: "credentials",
		APIHost:             "http://localhost:8080/api/",
		LogFilename:         "links-manager-client.log",
		StorageName:         "lmc.db",
	}
	userCredentials, err := setup(config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	auth := Auth{}
	auth.Config = config
	auth.UserCredentials = userCredentials

	storage, err := NewStorage(config.StoragePath())
	if err != nil {
		fmt.Printf("Storage opening failed %s", err.Error())
	}

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
			logger.Printf("JobResult received: %v", jobResult)
			if !jobResult.IsDone() {
				logger.Printf("job #%s is not done: %s\n", res.GetJobID(), res.(JobResult).lastError.Error())
				// Send signal, connection failed, so we need to stop send requests and wait reestablishing connection.
				if jobResult.ConnectionFailed() {
					noConnection <- true
				}
			} else {
				storage.Remove(res.GetJobID())
				logger.Printf("job #%s successed\n", res.GetJobID())
			}
		default:
			logger.Println("unknown result type")
		}
	})
	scheduler.Run()

	// Read previously saved uncompleted jobs from file
	readAllSavedJobsAndSchedule(scheduler, storage)

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
	go schedule(&auth, scheduler, noConnection, jobs, storage)

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
}

func readAllSavedJobsAndSchedule(scheduler *s.JobsScheduler, storage Storage) {
	savedJobs := []Job{}
	savedJobsData, err := storage.ReadAll()
	if err != nil {
		fmt.Println("Cannot read uncompleted jobs from storage %s", err.Error())
	}
	for i := 0; i < len(savedJobsData); i++ {
		savedJob := Job{}
		json.Unmarshal(savedJobsData[i], &savedJob)
		savedJobs = append(savedJobs, savedJob)
	}

	// Schedule uncompleted jobs
	for _, v := range savedJobs {
		err = scheduler.Add(v)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func schedule(auth *Auth, scheduler *s.JobsScheduler, noConnection chan bool, jobs chan Job, storage Storage) {
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
			connectionFailed = false
			readAllSavedJobsAndSchedule(scheduler, storage)
		case job := <-jobs:
			// Save job to storage, in case connection failed, we could restart jobs
			b, err := json.Marshal(job)
			if err != nil {
				// TODO write to log file
			} else {
				storage.Put(job.ID, b)
			}
			if !connectionFailed {
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
