/** ****************************************************************************************************************** **
    Workiz API wrapper
    written for GoLang
    Created 2022-06-20 by Nathan Thomas 
    Courtesy of BeelineRoutes.com

    current docs in v1
    https://developer.workiz.com/

** ****************************************************************************************************************** **/

package workiz 

import (
    "github.com/pkg/errors"

    // "fmt"
    "context"
    "strings"
    "time"
    "net/http"
    "encoding/json"
    "os"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

const apiURL = "https://api.workiz.com/api/v1"

var (
    ErrUnexpected       = errors.New("idk...")
	ErrNotFound 		= errors.New("Item was not found")
	ErrTooManyRecords	= errors.New("Too many records returned")
    ErrAuthExpired      = errors.New("Auth Expired")
    ErrQuota            = errors.New("Too many requests - quota limit")
)

type assignCrew func (context.Context, string, string, string, string) error 
type unassignCrew func (context.Context, string, string, string, string) error 

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Config struct {
    Token, Secret string 
}

func (this Config) Valid () bool {
    if len(this.Token) < 20 { return false } // i'm making these 20 so the example_config comes back as false
    if len(this.Secret) < 20 { return false }

    return true 
}

type baseResp struct {
    Flag, Error bool 
    Msg string 
    Data []struct {
        UUID, Client_id string 
    }
    Code int 
}

type baseRespDetails1 struct {
    Details struct {
        Error string 
    }
}

type baseRespDetails2 struct {
    Details []baseRespDetails1
}

type apiResp struct {
    baseResp
    baseRespDetails1
    
}

func (this *apiResp) UnmarshalJSON (b []byte) error {
    err := json.Unmarshal (b, &this.baseResp)
    if err != nil { return err }

    // now try the details
    one := baseRespDetails1{}
    err = json.Unmarshal (b, &one)
    if err == nil {
        this.baseRespDetails1 = one
        return nil  
    }

    // didn't work, try 2
    two := baseRespDetails2{}
    err = json.Unmarshal (b, &two)
    if err == nil {
        if len(two.Details) > 0 {
            this.baseRespDetails1 = two.Details[0]
        }
        return nil  
    }

    return err // this really didn't work
}

type workizTime struct {
    time.Time 
}

func (this *workizTime) UnmarshalJSON (b []byte) (err error) {
    s := strings.Trim(string(b), "\"")
    if s == "null" {
       this.Time = time.Time{} // go with an enpty date/time 
       return
    }

    this.Time, err = time.Parse("2006-01-02 15:04:05", s)
    return
}

type workizComment []string

func (this *workizComment) UnmarshalJSON (b []byte) error {

    // see if it's an empty string, if so, we're done
    if string(b) == `""` { return nil }

    // we have comments
    var data []struct {
        Comment string 
    }

    err := json.Unmarshal (b, &data)
    if err != nil { return err } // this is bad

    // this worked, so populate our comments
    for _, c := range data {
        *this = append (*this, c.Comment)
    }

    return nil 
}

type teamGeneric struct {
    Id, Name string
}

//----- ERRORS ---------------------------------------------------------------------------------------------------------//
type Error struct {
	Msg string
	StatusCode int
}

func (this *Error) Err () error {
	if this == nil { return nil } // no error
	switch this.StatusCode {
	case http.StatusUnauthorized:
        return errors.Wrapf (ErrAuthExpired, "Unauthorized : %d : %s", this.StatusCode, this.Msg)
	}
	// just a default
	return errors.Wrapf (ErrUnexpected, "Workiz Error : %d : %s", this.StatusCode, this.Msg)
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CLASS -----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Workiz struct {
	
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// this is just used for local testing
// so you don't have to keep your actual tokens in the repo
func parseConfig (jsonFile string) (*Config, error) {
	config, err := os.Open(jsonFile)
	if err != nil { return nil, errors.WithStack (err) }

	jsonParser := json.NewDecoder (config)

    ret := &Config{}
	err = jsonParser.Decode (ret)
    return ret, errors.WithStack(err)
}

/* handles the high level logic of changing which crew members are assigned to a job or lead
crew members need to be assigned one at a time
and if you assign the same one twice, you get an error
so we need to get the currently assigned ones first, then figure out if more need to be added or removed

this is brutal, the name of the crew member you get from the get job call may not actually match the crew member's current name
so in order to remove them, and add them, we need to reference them by the id first, and then their current name
my guess is they don't use a relational database, so if you change the crew member's name after assigning them to a job it stays
as the old name in the job table/object
*/
func (this *Workiz) handleCrew (ctx context.Context, existingTeam []*teamGeneric, token, secret, jobId string, team Members, fullNames []string, assFn assignCrew, unassFn unassignCrew) error {
    // first step, add the missing ones
    for _, name := range fullNames {
        nameId := team.FindId (name) // find the id by the name
        if len(nameId) == 0 { continue } // this is a bad name, no id, so nothing we can do at this point

        exists := false 
        for _, team := range existingTeam {
            //if strings.EqualFold (team.Name, name) { 
            if strings.EqualFold (team.Id, nameId) {
                exists = true 
                break 
            }
        }

        if exists == false {
            // it's missing so add it
            err := assFn (ctx, token, secret, jobId, name)
            if err != nil { return err }
        }
    }

    // second step, remove the assigned crew that are no longer assigned
    for _, existing := range existingTeam {
        existingName := team.FindName (existing.Id) // get the actual current name using the assigned id
        if len(existingName) == 0 { continue } // not sure this can happen, but we can't do anything if it does

        exists := false 
        for _, name := range fullNames {
            if strings.EqualFold (existingName, name) {
                exists = true 
                break
            }
        }

        if exists == false {
            // they're currently assigned and we need to remove them
            err := unassFn (ctx, token, secret, jobId, existingName)
            if err != nil { return err }
        }
    }
    return nil // we got everything figured out
}
