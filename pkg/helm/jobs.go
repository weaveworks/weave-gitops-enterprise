package helm

import (
	"time"

	"github.com/cheshir/ttlcache"
	"github.com/google/uuid"
)

// JobTTL is how long until a job expires
const JobTTL = time.Hour

// JobCleanupInterval is how often the jobs queue is checked for expired jobs
const JobCleanupInterval = time.Hour

type Jobs struct {
	results ttlcache.Cache
}

type JobResult struct {
	Result string
	Error  error
}

// NewJobs creates an in memory job cache
// Jobs expire after after a hour
func NewJobs() *Jobs {
	return &Jobs{
		results: *ttlcache.New(JobCleanupInterval),
	}
}

// New creates and saves a new empty job and returns its id
func (j *Jobs) New() string {
	nextId := uuid.New().String()
	j.Set(nextId, JobResult{})
	return nextId
}

// Get returns the job result and true if the job exists
func (j *Jobs) Get(id string) (JobResult, bool) {
	res, found := j.results.Get(ttlcache.StringKey(id))
	if !found {
		return JobResult{}, false
	}

	return res.(JobResult), found
}

// Set sets the job result
func (j *Jobs) Set(id string, result JobResult) {
	j.results.Set(ttlcache.StringKey(id), result, JobTTL)
}
