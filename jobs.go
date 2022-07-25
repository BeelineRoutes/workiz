/** ****************************************************************************************************************** **
	Calls related to jobs

    
** ****************************************************************************************************************** **/

package workiz 

import (
    "github.com/gofrs/uuid"
    "github.com/pkg/errors"
    
    "fmt"
    "net/http"
    "net/url"
    "context"
    "time"
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

type Job struct {
    UUID *uuid.UUID 
    SerialId, ClientId int 
    JobDateTime, JobEndDateTime, CreatedDate, PaymentDueDate, LastStatusUpdate time.Time
    JobTotalPrice, JobAmountDue, SubTotal, SubStatus, JobType, ReferralCompany, Timezone, ServiceArea string 
    Phone, PhoneExt, Email, Comments, FirstName, LastName, Company, JobNotes, JobSource, CreatedBy string 
    Address, City, State, PostalCode, Country, Unit string 
    Latitude, Longitude string 
    ItemCost string `json:"item_cost"`
    TechCost string `json:"tech_cost"`
    Status JobStatus
    Team []struct {
        Id string `json:"id"`
        Name string `json:"name"`
    }
}

type CreateJob struct {
    AuthSecret string `json:"auth_secret"`
    JobDateTime time.Time 
    ClientId int
    Phone, Email, FirstName, LastName, Address, City, State, PostalCode string 
    JobType, ReferralCompany, Timezone, JobSource, JobNotes, CreatedBy, ServiceArea string 
}

type jobResponse []struct {
    Flag bool 
    Data Job
}

// takes the jobs out of whatever this parent object is for
func (this jobResponse) toJobs () (ret []*Job) {
    for _, j := range this {
        if j.Flag {
            ret = append (ret, &j.Data)
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
    
    errObj, err := this.send (ctx, http.MethodGet, token, fmt.Sprintf("jobs/get/%s", jobId), nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj } // something else bad

    jobs := resp.toJobs() // pull out the jobs
    if len(jobs) == 0 {
        return nil, errors.Wrap (ErrNotFound, jobId)
    } else if len(jobs) > 1 {
        return nil, errors.Wrapf (ErrUnexpected, "More than 1 job found for id '%s'", jobId)
    }

    // we're here, we're good
    return jobs[0], nil
}

// returns all jobs that match our conditions
func (this *Workiz) ListJobs (ctx context.Context, token string, start time.Time, status ...JobStatus) ([]*Job, error) {
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

    for i := 0; i < 100; i++ { // stay in a loop as long as we're pulling jobs
        params.Set("offset", fmt.Sprintf("%d", i)) // set our next page
        var resp jobResponse
        
        errObj, err := this.send (ctx, http.MethodGet, token, fmt.Sprintf("jobs/all?%s", params.Encode()), nil, &resp)
        if err != nil { return nil, errors.WithStack(err) } // bail
        if errObj != nil { return nil, errObj } // something else bad

        // we're here, we're good
        newJobs := resp.toJobs()
        ret = append (ret, newJobs...)
        
        // 100 is the default records count from above
        if len(newJobs) >= 100 { return ret, nil } // we finished
    }
    return ret, errors.Errorf ("received over %d jobs in your history", len(ret))
}

// sets the list of all assigned employees for a job
// only updates the list of employees assigned to a job
func (this *Workiz) UpdateJob (ctx context.Context, token, secret, jobId string, start time.Time, employeeIds ...string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID string 
        JobDateTime time.Time 
    }
    data.UUID = jobId 
    data.AuthSecret = secret
    data.JobDateTime = start 

    errObj, err := this.send (ctx, http.MethodPost, token, "jobs/update", data, nil)
    if err != nil { return errors.WithStack(err) } // bail
    if errObj != nil { return errObj } // something else bad
    
    // we're here, we're good
    return nil
}

// creates a new job in the system
func (this *Workiz) CreateJob (ctx context.Context, token, secret string, job *CreateJob) error {
    job.AuthSecret = secret
    
    errObj, err := this.send (ctx, http.MethodPost, token, "jobs/create", job, nil)
    if err != nil { return errors.WithStack(err) } // bail
    if errObj != nil { return errObj } // something else bad
    
    // we're here, we're good
    return nil
}
