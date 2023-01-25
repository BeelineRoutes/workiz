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
    "strings"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Lead struct {
    UUID string
    SerialId, ClientId int 
    LeadDateTime, LeadEndDateTime, CreatedDate, PaymentDueDate, LastStatusUpdate time.Time
    LeadTotalPrice, LeadAmountDue, SubTotal, SubStatus, LeadType, ReferralCompany, Timezone, ServiceArea string 
    Phone, PhoneExt, SecondPhone, Email, Comments, FirstName, LastName, Company, LeadNotes, LeadSource, CreatedBy string 
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
    Comments, JobType, JobSource, LeadNotes, ServiceArea string
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
    
    err := this.send (ctx, http.MethodGet, token, fmt.Sprintf("lead/get/%s/", leadId), nil, &resp)
    if err != nil { return nil, err } // bail
    
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
        
        err := this.send (ctx, http.MethodGet, token, fmt.Sprintf("lead/all/?%s", params.Encode()), nil, &resp)
        if err != nil { return nil, err } // bail
        
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

    err := this.send (ctx, http.MethodPost, token, "lead/update/", data, nil)
    if err != nil { return err } // bail
    
    // we're here, we're good
    return nil
}

// handles the high level logic of changing which crew members are assigned to a lead
// crew members need to be assigned one at a time
// and if you assign the same one twice, you get an error
// so we need to get the currently assigned ones first, then figure out if more need to be added or removed
func (this *Workiz) UpdateLeadCrew (ctx context.Context, token, secret, leadId string, fullNames []string) error {
    existing, err := this.GetLead (ctx, token, leadId)
    if err != nil { return err }

    // first step, add the missing ones
    for _, name := range fullNames {
        exists := false 
        for _, team := range existing.Team {
            if strings.EqualFold (team.Name, name) { 
                exists = true 
                break 
            }
        }

        if exists == false {
            // it's missing so add it
            err = this.AssignLeadCrew (ctx, token, secret, leadId, name)
            if err != nil { return err }
        }
    }

    // second step, remove the assigned crew that are no longer assigned
    for _, team := range existing.Team {
        exists := false 
        for _, name := range fullNames {
            if strings.EqualFold (team.Name, name) {
                exists = true 
                break
            }
        }

        if exists == false {
            // they're currently assigned and we need to remove them
            err = this.UnassignLeadCrew (ctx, token, secret, leadId, team.Name)
            if err != nil { return err }
        }
    }
    return nil // we got everything figured out
}

// assigns a lead to the crew names
func (this *Workiz) AssignLeadCrew (ctx context.Context, token, secret, leadId string, fullName string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, User string 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    data.User = fullName
    
    return this.send (ctx, http.MethodPost, token, "lead/assign/", data, nil)
}

// unassigns a lead to the crew names
func (this *Workiz) UnassignLeadCrew (ctx context.Context, token, secret, leadId string, fullName string) error {
    var data struct {
        AuthSecret string `json:"auth_secret"`
        UUID, User string 
    }
    data.UUID = leadId 
    data.AuthSecret = secret
    data.User = fullName // it's based on name, not id
    
    return this.send (ctx, http.MethodPost, token, "lead/unassign/", data, nil)
}

// creates a new lead in the system
// returns the uuid of the newly created lead, so we can then assign crew members
func (this *Workiz) CreateLead (ctx context.Context, token, secret string, lead *CreateLead) (string, error) {
    lead.AuthSecret = secret

    // we need the id right away
    resp := &apiResp{}
    
    err := this.send (ctx, http.MethodPost, token, "lead/create/", lead, resp)
    if err != nil { return "", err } // bail
    
    if resp.Flag == false || len(resp.Data) == 0 {
        return "", errors.Errorf ("didn't get expected data back from creating a lead: %+v", resp)
    }
    
    // we're here, we're good
    return resp.Data[0].UUID, nil
}
