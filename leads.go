/** ****************************************************************************************************************** **
	Calls related to leads (estimates)

    
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

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Lead struct {
    UUID *uuid.UUID 
    SerialId, ClientId int 
    LeadDateTime, LeadEndDateTime, CreatedDate, PaymentDueDate, LastStatusUpdate time.Time
    LeadTotalPrice, LeadAmountDue, SubTotal, SubStatus, LeadType, ReferralCompany, Timezone, ServiceArea string 
    Phone, PhoneExt, Email, Comments, FirstName, LastName, Company, LeadNotes, LeadSource, CreatedBy string 
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

type CreateLead struct {
    AuthSecret string `json:"auth_secret"`
    LeadDateTime, LeadEndDateTime time.Time 
    ClientId int
    Phone, Email, FirstName, LastName, Address, City, State, PostalCode string 
    Comments, JobType, JobSource, LeadNotes string
}

type leadResponse []struct {
    Flag bool 
    Data Lead
}

// takes the leads out of whatever this parent object is for
func (this leadResponse) toLeads () (ret []*Lead) {
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

// gets the info about a specific lead
func (this *Workiz) GetLead (ctx context.Context, token, leadId string) (*Lead, error) {
    var resp leadResponse
    
    errObj, err := this.send (ctx, http.MethodGet, token, fmt.Sprintf("lead/get/%s/", leadId), nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj } // something else bad

    leads := resp.toLeads() // pull out the leads
    if len(leads) == 0 {
        return nil, errors.Wrap (ErrNotFound, leadId)
    } else if len(leads) > 1 {
        return nil, errors.Wrapf (ErrUnexpected, "More than 1 lead found for id '%s'", leadId)
    }

    // we're here, we're good
    return leads[0], nil
}

// returns all leads that match our conditions
func (this *Workiz) ListLeads (ctx context.Context, token string, start time.Time, status ...JobStatus) ([]*Lead, error) {
    ret := make([]*Lead, 0) // main list to return
    
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

    for i := 0; i < 10; i++ { // stay in a loop as long as we're pulling leads
        params.Set("offset", fmt.Sprintf("%d", i)) // set our next page
        var resp leadResponse
        
        errObj, err := this.send (ctx, http.MethodGet, token, fmt.Sprintf("lead/all/?%s", params.Encode()), nil, &resp)
        if err != nil { return nil, errors.WithStack(err) } // bail
        if errObj != nil { return nil, errObj } // something else bad

        // we're here, we're good
        newLeads := resp.toLeads()
        ret = append (ret, newLeads...)
        
        // 100 is the default records count from above
        if len(newLeads) < 100 { return ret, nil } // we finished
    }
    return ret, errors.Wrapf (ErrTooManyRecords, "received over %d leads in your history", len(ret))
}

// updates the start/end time for a lead at UTC
func (this *Workiz) UpdateLeadSchedule (ctx context.Context, token, secret, leadId string, startTime time.Time, duration time.Duration) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, Timezone string 
        LeadDateTime, LeadEndDateTime time.Time 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    data.Timezone = "UTC" // we're always in utc
    data.LeadDateTime = startTime 
    data.LeadEndDateTime = data.LeadDateTime.Add(duration)

    errObj, err := this.send (ctx, http.MethodPost, token, "lead/update/", data, nil)
    if err != nil { return errors.WithStack(err) } // bail
    if errObj != nil { return errObj } // something else bad
    
    // we're here, we're good
    return nil
}

// assigns a lead to the crew names
func (this *Workiz) AssignLeadCrew (ctx context.Context, token, secret, leadId string, fullNames []string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, User string 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    
    for _, name := range fullNames {
        data.User = name // it's based on name, not id

        errObj, err := this.send (ctx, http.MethodPost, token, "lead/assign/", data, nil)
        if err != nil { return errors.WithStack(err) } // bail
        if errObj != nil { return errObj } // something else bad
    }
    
    // we're here, we're good
    return nil
}

// unassigns a lead to the crew names
func (this *Workiz) UnassignLeadCrew (ctx context.Context, token, secret, leadId string, fullNames []string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, User string 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    
    for _, name := range fullNames {
        data.User = name // it's based on name, not id

        errObj, err := this.send (ctx, http.MethodPost, token, "lead/unassign/", data, nil)
        if err != nil { return errors.WithStack(err) } // bail
        if errObj != nil { return errObj } // something else bad
    }
    
    // we're here, we're good
    return nil
}

// creates a new lead in the system
func (this *Workiz) CreateLead (ctx context.Context, token, secret string, lead *CreateLead) error {
    lead.AuthSecret = secret
    lead.Timezone = "UTC" // we're always in utc
    
    errObj, err := this.send (ctx, http.MethodPost, token, "lead/create/", lead, nil)
    if err != nil { return errors.WithStack(err) } // bail
    if errObj != nil { return errObj } // something else bad
    
    // we're here, we're good
    return nil
}
