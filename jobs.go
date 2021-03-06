package main

import "net/http"

// Job test job, which should implements required method GetID() string
type Job struct {
	ID   string
	Link *Link
}

// JobResult test job result, which should implements required methods
// IsDone() and IsCorrupted(). The last returns true if processing failed
// and no need to restart the job.
type JobResult struct {
	lastError error
	job       Job
}

// GetID implement Job interface
func (job Job) GetID() string {
	return job.ID
}

// GetJobID implements required JobResult interface to use with jobs-scheduler
func (jobResult JobResult) GetJobID() string {
	return jobResult.job.GetID()
}

// IsDone implement JobResult interface
func (jobResult JobResult) IsDone() bool {
	return jobResult.lastError == nil
}

// IsCorrupted implement JobResult interface
func (jobResult JobResult) IsCorrupted() bool {
	return false
}

// ConnectionFailed indicates if connection failed or service unavailable. In both cases need to retry the job.
func (jobResult JobResult) ConnectionFailed() bool {
	res := false
	switch t := jobResult.lastError.(type) {
	case *APIConnectionFailed:
		res = true
	case *APIError:
		if t.code >= http.StatusInternalServerError {
			res = true
		}
	}

	return res
}
