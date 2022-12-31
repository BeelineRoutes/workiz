
package workiz 

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"context"
	"time"
)

func TestJobGet (t *testing.T) {
	w := &Workiz{}
	cfg := getRealConfig(t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of jobs, only unscheduled ones
	job, err := w.GetJob (ctx, cfg.Token, "OWX12J")
	if err != nil { t.Fatal (err) }

	assert.Equal (t, "OWX12J", job.UUID, "not filled in")
	assert.Equal (t, 1, len(job.Comments))
	assert.Equal (t, "this is a note, not a comment", job.Comments[0])
	
	/*
	for _, j := range jobs {
		t.Logf ("%+v\n", j)
	}
	*/
}

func TestJobs (t *testing.T) {
	w := &Workiz{}
	cfg := getRealConfig(t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of jobs, only unscheduled ones
	jobs, err := w.ListJobs (ctx, cfg.Token, time.Now(), JobStatus_submitted)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(jobs) > 0, "expecting at least 1 job")
	assert.NotEqual (t, "", jobs[0].UUID, "not filled in")
	assert.NotEqual (t, "", jobs[0].ClientId, "not filled in")
	assert.NotEqual (t, "", jobs[0].Address, "not filled in")
	
	/*
	for _, j := range jobs {
		t.Logf ("%+v\n", j)
	}
	*/
}

