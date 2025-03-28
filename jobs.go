/** ****************************************************************************************************************** **
	Calls related to jobs

    
** ****************************************************************************************************************** **/

package workiz 

import (
    "github.com/pkg/errors"
    
    "fmt"
    "net/http"
    "net/url"
    "context"
    "time"
    "encoding/json"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type JobStatus string 

const (
	JobStatus_submitted         = JobStatus("Submitted")
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// Workiz has decided to return either an int or a string depending on what the user entered for the unit, so I need to try both
type Unit struct {
	Value string
}

// Implement UnmarshalJSON to handle both string and int
func (this *Unit) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal as a string
	if err := json.Unmarshal(data, &this.Value); err == nil {
		return nil
	}

	// If it's not a string, then try it as an int
	var i int
	if err := json.Unmarshal(data, &i); err == nil && i != 0 {
		this.Value = fmt.Sprintf("%d", i)
		return nil
	}

	return nil // i'm just going to ignore an error if this still doesn't work
}

type Job struct {
    UUID string
    SerialId, ClientId int 
    JobDateTime, JobEndDateTime, CreatedDate, PaymentDueDate, LastStatusUpdate workizTime
    JobTotalPrice, JobAmountDue, SubTotal json.Number
    SubStatus, JobType, ReferralCompany, Timezone, ServiceArea string 
    Phone, PhoneExt, SecondPhone, Email, FirstName, LastName, Company, JobNotes, JobSource, CreatedBy string 
    Address, City, State, PostalCode, Country string 
    Unit Unit
    Latitude, Longitude float64 
    ItemCost int `json:"item_cost"`
    TechCost int `json:"tech_cost"`
    Status JobStatus
    Team []struct {
        Id int `json:"id"`
        Name string `json:"name"`
    }
    Comments workizComment
}

func (this *Job) toGeneric () (ret []*teamGeneric) {
    for _, t := range this.Team {
        ret = append(ret, &teamGeneric {
            Id: fmt.Sprintf("%d", t.Id),
            Name: t.Name,
        })
    }
    return 
}

type baseAuth struct {
    AuthSecret string `json:"auth_secret"`
}

type CreateJob struct {
    baseAuth
    JobDateTime, JobEndDateTime time.Time 
    ClientId int
    Phone, Email, FirstName, LastName, Address, City, State, PostalCode string 
    JobType, JobSource, JobNotes, ServiceArea string 
}

type jobResponse struct {
    Flag, Has_more bool 
    Data []*Job
}

// takes the jobs out of whatever this parent object is for
// we don't have great control over the time range for jobs
func (this jobResponse) toJobs (start, end time.Time) (ret []*Job) {
    for _, job := range this.Data {
        if start.IsZero() || (job.JobDateTime.After(start) && job.JobDateTime.Before(end)) {
            ret = append (ret, job)
        }
    }
    return
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -----------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// gets the info about a specific job
func (this *Workiz) GetJob (ctx context.Context, token, jobId string) (*Job, error) {
    var resp jobResponse
    
    err := this.send (ctx, 0, http.MethodGet, token, fmt.Sprintf("job/get/%s/", jobId), nil, &resp)
    if err != nil { return nil, err } // bail
    
    jobs := resp.toJobs(time.Time{}, time.Time{}) // pull out the jobs
    if len(jobs) == 0 {
        return nil, errors.Wrap (ErrNotFound, jobId)
    } else if len(jobs) > 1 {
        return nil, errors.Wrapf (ErrUnexpected, "More than 1 job found for id '%s'", jobId)
    }

    // we're here, we're good
    return jobs[0], nil
}

// returns all jobs that match our conditions
func (this *Workiz) ListJobs (ctx context.Context, token string, start, end time.Time, status ...JobStatus) ([]*Job, error) {
    ret := make([]*Job, 0) // main list to return
    
    params := url.Values{}
    params.Set("records", "100") // docs say 100 is the most you can request at a time
    if len(status) == 0 {
        params.Set("only_open", "true") // default
    } else {
        params.Set("only_open", "false")
    }

    for _, stat := range status {
        params.Add("status", string(stat))
    }
    
    if start.IsZero() == false {
        params.Set("start_date", start.Format("2006-01-02"))
    }

    // because unscheduled jobs come in with scheduled ones, we need to compare their ids to make sure we don't include unscheduled ones
    uMap := make(map[string]bool)
    unscheduled, _ := this.ListUnscheduledJobs (ctx, token)
    for _, u := range unscheduled {
        uMap[u.UUID] = true
    }

    for i := 0; i < 10; i++ { // stay in a loop as long as we're pulling jobs
        params.Set("offset", fmt.Sprintf("%d", i)) // set our next page
        var resp jobResponse
        
        err := this.send (ctx, 0, http.MethodGet, token, fmt.Sprintf("job/all/?%s", params.Encode()), nil, &resp)
        if err != nil { return nil, err } // bail
        
        // we're here, we're good
        newJobs := resp.toJobs(start, end)
        
        if len(newJobs) == 0 {
            // means we didn't pull any more jobs from within our date range
            return ret, nil 
        }

        for _, nj := range newJobs {
            if uMap[nj.UUID] == false {
                ret = append (ret, nj)
            }
        }

        if resp.Has_more == false { return ret, nil } // we're done
    }
    return ret, errors.Wrapf (ErrTooManyRecords, "received over %d jobs in your history. %s - %s", len(ret), start, end)
}

// lists the unscheduled jobs, which still have a job date and time... :shrug:
func (this *Workiz) ListUnscheduledJobs (ctx context.Context, token string) ([]*Job, error) {
    ret := make([]*Job, 0) // main list to return
    
    params := url.Values{}
    params.Set("records", "100") // docs say 100 is the most you can request at a time
    params.Set("only_open", "true") // default
    
    for i := 0; i < 10; i++ { // stay in a loop as long as we're pulling jobs
        params.Set("offset", fmt.Sprintf("%d", i)) // set our next page
        var resp jobResponse
        
        err := this.send (ctx, 0, http.MethodGet, token, fmt.Sprintf("job/all/?%s", params.Encode()), nil, &resp)
        if err != nil { return nil, err } // bail
        
        // we're here, we're good
        newJobs := resp.toJobs(time.Time{}, time.Time{})
        
        if len(newJobs) == 0 {
            // means we didn't pull any more jobs from within our date range
            return ret, nil 
        }

        ret = append (ret, newJobs...)

        if resp.Has_more == false { return ret, nil } // we're done
    }
    return ret, errors.Wrapf (ErrTooManyRecords, "received over %d unscheduled jobs in your history.", len(ret))
}

// updates the start/end time for a job
// the time needs to be set to whatever timezone the user's account is in
func (this *Workiz) UpdateJobSchedule (ctx context.Context, token, secret, jobId string, startTime time.Time, duration time.Duration) error {
    var data struct {
        baseAuth
        UUID, Timezone string 
        JobDateTime, JobEndDateTime time.Time 
    }
    data.UUID = jobId 
    data.AuthSecret = secret
    data.JobDateTime = startTime 
    data.JobEndDateTime = data.JobDateTime.Add(duration)

    err := this.send (ctx, 0, http.MethodPost, token, "job/update/", data, nil)
    if err != nil { return err } // bail
    
    // we're here, we're good
    return nil
}

// wrapper around our reusable assiging crew function
// just give it the correct assign and unassign functions
func (this *Workiz) UpdateJobCrew (ctx context.Context, token, secret, jobId string, team Members, fullNames []string) error {
    existing, err := this.GetJob (ctx, token, jobId)
    if err != nil { return err }

    return this.handleCrew (ctx, existing.toGeneric(), token, secret, jobId, team, fullNames, this.AssignJobCrew, this.UnassignJobCrew)
}

// assigns a job to the crew names
func (this *Workiz) AssignJobCrew (ctx context.Context, token, secret, jobId string, fullName string) error {
    var data struct {
        baseAuth
        UUID, User string 
    }
    data.UUID = jobId 
    data.AuthSecret = secret
    data.User = fullName // it's based on name, not id

    err := this.send (ctx, 0, http.MethodPost, token, "job/assign/", data, nil)
    if err != nil { return err } // bail
    
    // we're here, we're good
    return nil
}

// unassigns a job to the crew names
func (this *Workiz) UnassignJobCrew (ctx context.Context, token, secret, jobId string, fullName string) error {
    var data struct {
        baseAuth
        UUID, User string 
    }
    data.UUID = jobId 
    data.AuthSecret = secret
    data.User = fullName // it's based on name, not id
    
    err := this.send (ctx, 0, http.MethodPost, token, "job/unassign/", data, nil)
    if err != nil { return err } // bail
    
    // we're here, we're good
    return nil
}

// creates a new job in the system
// jobs are created in the timezone of the account. so if we have the JobDateTime: "2022-12-18 15:00:00" it will create the job at 3pm est
// so we need to convert this time from UTC to the local timezone for the account
func (this *Workiz) CreateJob (ctx context.Context, token, secret string, job *CreateJob) (string, error) {
    job.AuthSecret = secret
    resp := &apiResp{}
    
    err := this.send (ctx, 0, http.MethodPost, token, "job/create/", job, resp)
    if err != nil { return "", err } // bail
    
    if resp.Flag == false || len(resp.Data) == 0 {
        return "", errors.Errorf ("didn't get expected data back from creating a lead: %+v", resp)
    }
    
    // we're here, we're good
    return resp.Data[0].UUID, nil
}

// all jobs need a job type
// this creates it if its missing
func (this *Workiz) CreateJobType (ctx context.Context, token, secret, jobType string) error {
    var data struct {
        baseAuth
        JobType string 
    }
    data.AuthSecret = secret
    data.JobType = jobType
    
    err := this.send (ctx, 0, http.MethodPost, token, "jobType/createIfNotExists/", data, nil)
    if err != nil { return err } // bail
    
    // we're here, we're good
    return nil
}
