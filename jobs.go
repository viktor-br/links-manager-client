package main

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

// ConnectionFailed indicates if connection failed.
func (jobResult JobResult) ConnectionFailed() bool {
	if e, ok := jobResult.lastError.(*APIError); ok {
		return e.code >= 400
	}
	return false
}
