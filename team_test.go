
package workiz 

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"context"
	"time"
)

func TestTeam (t *testing.T) {
	w := &Workiz{}
	cfg := getRealConfig(t)

	ctx, cancel := context.WithTimeout (context.Background(), time.Minute) // this should take < 1 minute
	defer cancel()

	// get our list of members, only unscheduled ones
	members, err := w.ListTeam (ctx, cfg.Token)
	if err != nil { t.Fatal (err) }

	assert.Equal (t, true, len(members) > 0, "expecting at least 1 team member")
	assert.NotEqual (t, "", members[0].Id, "not filled in")
	assert.NotEqual (t, "", members[0].Name, "not filled in")
	
	/*
	for _, j := range members {
		t.Logf ("%+v\n", j)
	}
	*/
}

