/** ****************************************************************************************************************** **
	Calls related to leads (estimates)

    
** ****************************************************************************************************************** **/

package workiz 

import (
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
    UUID, SerialId, ClientId string
    LeadDateTime, LeadEndDateTime, CreatedDate, PaymentDueDate, LastStatusUpdate workizTime
    LeadTotalPrice, LeadAmountDue, SubTotal, SubStatus, LeadType, ReferralCompany, Timezone, ServiceArea string 
    Phone, PhoneExt, SecondPhone, Email, Comments, FirstName, LastName, Company, LeadNotes, LeadSource, CreatedBy string 
    Address, City, State, PostalCode, Country string 
    Unit Unit
    Latitude, Longitude string 
    ItemCost string `json:"item_cost"`
    TechCost string `json:"tech_cost"`
    Status JobStatus
    Team []struct {
        Id string `json:"id"`
        Name string `json:"name"`
    }
}

func (this *Lead) toGeneric () (ret []*teamGeneric) {
    for _, t := range this.Team {
        ret = append(ret, &teamGeneric {
            Id: t.Id,
            Name: t.Name,
        })
    }
    return 
}

type CreateLead struct {
    AuthSecret string `json:"auth_secret"`
    LeadDateTime, LeadEndDateTime time.Time 
    ClientId int
    Phone, Email, FirstName, LastName, Address, City, State, PostalCode string 
    Comments, JobType, JobSource, LeadNotes, ServiceArea string
}

type leadResponse struct {
    Flag bool 
    Has_more bool 
    Data []*Lead
}

func (this leadResponse) toJobs (start, end time.Time) (ret []*Lead) {
    for _, lead := range this.Data {
        if start.IsZero() || // we're looking for unscheduled ones
            (lead.LeadDateTime.After(start) && lead.LeadDateTime.Before(end)) {
            ret = append (ret, lead)
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
    resp := &leadResponse{}
    
    err := this.send (ctx, 0, http.MethodGet, token, fmt.Sprintf("lead/get/%s/", leadId), nil, resp)
    if err != nil { return nil, err } // bail
    
    if len(resp.Data) == 0 {
        return nil, errors.Wrap (ErrNotFound, leadId)
    } else if len(resp.Data) > 1 {
        return nil, errors.Wrapf (ErrUnexpected, "More than 1 lead found for id '%s'", leadId)
    }

    // we're here, we're good
    return resp.Data[0], nil
}

// returns all leads that match our conditions
func (this *Workiz) ListLeads (ctx context.Context, token string, start, end time.Time, status ...JobStatus) ([]*Lead, error) {
    ret := make([]*Lead, 0) // main list to return
    
    params := url.Values{}
    params.Set("records", "100") // docs say 100 is the most you can request at a time
    
    for _, stat := range status {
        params.Add("status", string(stat))
    }
    
    if start.IsZero() == false {
        params.Set("start_date", start.Format("2006-01-02"))
    }

    for i := 0; i < 10; i++ { // stay in a loop as long as we're pulling leads
        params.Set("offset", fmt.Sprintf("%d", i)) // set our next page
        resp := &leadResponse{}
        
        err := this.send (ctx, 0, http.MethodGet, token, fmt.Sprintf("lead/all/?%s", params.Encode()), nil, resp)
        if err != nil { return nil, err } // bail
        
        // we're here, we're good
        leads := resp.toJobs (start, end) // use this to filter out leads outside of the date range

        if len(leads) == 0 {
            // we're done, all these are in the future
            // i do'nt know if this is correct, i'm just hoping they're ordered by target date
            // they're probably ordered by created, but this works close enough
            return ret, nil 
        }

        // add them to our list
        ret = append (ret, leads...)
        
        if resp.Has_more == false { return ret, nil } // we finished
    }
    return ret, errors.Wrapf (ErrTooManyRecords, "received over %d leads in your history. %s - %s", len(ret), start, end)
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

    err := this.send (ctx, 0, http.MethodPost, token, "lead/update/", data, nil)
    if err != nil { return err } // bail
    
    // we're here, we're good
    return nil
}

// wrapper around our re-usable assign function, which is super complicated unfortuantely 
func (this *Workiz) UpdateLeadCrew (ctx context.Context, token, secret, leadId string, team Members, fullNames []string) error {
    existing, err := this.GetLead (ctx, token, leadId)
    if err != nil{ return err }

    return this.handleCrew (ctx, existing.toGeneric(), token, secret, leadId, team, fullNames, this.AssignLeadCrew, this.UnassignLeadCrew)
}

// assigns a lead to the crew names
func (this *Workiz) AssignLeadCrew (ctx context.Context, token, secret, leadId, fullName string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, User string 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    data.User = fullName
    
    return this.send (ctx, 0, http.MethodPost, token, "lead/assign/", data, nil)
}

// unassigns a lead to the crew names
func (this *Workiz) UnassignLeadCrew (ctx context.Context, token, secret, leadId, fullName string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, User string 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    data.User = fullName // it's based on name, not id
    
    return this.send (ctx, 0, http.MethodPost, token, "lead/unassign/", data, nil)
}

// creates a new lead in the system
// returns the uuid of the newly created lead, so we can then assign crew members
func (this *Workiz) CreateLead (ctx context.Context, token, secret string, lead *CreateLead) (string, error) {
    lead.AuthSecret = secret

    // we need the id right away
    resp := &apiResp{}
    
    err := this.send (ctx, 0, http.MethodPost, token, "lead/create/", lead, resp)
    if err != nil { return "", err } // bail
    
    if resp.Flag == false || len(resp.Data) == 0 {
        return "", errors.Errorf ("didn't get expected data back from creating a lead: %+v", resp)
    }
    
    // we're here, we're good
    return resp.Data[0].UUID, nil
}
