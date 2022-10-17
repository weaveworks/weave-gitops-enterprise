package helm

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cheshir/ttlcache"
)

type JobResult struct {
	Result string
	Error  error
}

type Jobs struct {
	ids     *int64
	results ttlcache.Cache
}

// constructor
func NewJobs() *Jobs {
	var ids int64
	return &Jobs{
		ids:     &ids,
		results: *ttlcache.New(time.Hour),
	}
}

func (j *Jobs) New() string {
	nextId := fmt.Sprint(atomic.AddInt64(j.ids, 1))
	j.Set(nextId, JobResult{})
	return nextId
}

func (j *Jobs) Get(id string) (JobResult, bool) {
	res, found := j.results.Get(ttlcache.StringKey(id))
	if !found {
		return JobResult{}, false
	}

	return res.(JobResult), found
}

func (j *Jobs) Set(id string, result JobResult) {
	j.results.Set(ttlcache.StringKey(id), result, time.Hour)
}
