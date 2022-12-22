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

    //"fmt"
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
)

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

type apiResp struct {
    Flag, Error bool 
    Msg string 
    Data []struct {
        UUID, Client_id string 
    }
    Details struct {
        Error string 
    }
    Code int 
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
	return errors.Wrapf (ErrUnexpected, "Workiz Error : %d : %s : %s", this.StatusCode, this.Msg)
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CLASS -----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Workiz struct {
	
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func ParseConfig (jsonFile string) (*Config, error) {
	config, err := os.Open(jsonFile)
	if err != nil { return nil, errors.WithStack (err) }

	jsonParser := json.NewDecoder (config)

    ret := &Config{}
	err = jsonParser.Decode (ret)
    return ret, errors.WithStack(err)
}
