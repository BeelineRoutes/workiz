/** ****************************************************************************************************************** **
	Calls related to clients (customers)

    
** ****************************************************************************************************************** **/

package workiz 

import (
    "github.com/pkg/errors"
    
    //"fmt"
    "net/http"
    "context"
    
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- CONSTS ----------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

type Client struct {
    AuthSecret string `json:"auth_secret"`
    Id, FirstName, LastName, Address, City, State, Zip, Source, Email string `json:",omitempty"`
    AllowBilling bool 
}

type getClientResp struct {
    Flag bool 
    Data Client 
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -----------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// creates a new client
func (this *Workiz) CreateClient (ctx context.Context, token, secret string, client *Client) error {
    resp := &apiResp{}
    client.AuthSecret = secret
    
    errObj, err := this.send (ctx, http.MethodPost, token, "Client/create/", client, resp)
    if err != nil { return errors.WithStack(err) } // bail
    if errObj != nil { return errObj.Err() } // something else bad

    if resp.Flag != true {
        return errors.Errorf("response flag was not true : %+v", resp)
    }

    client.Id = resp.Data[0].Client_id // copy this over

    return nil // we're good
}

// get a single client by the id
func (this *Workiz) GetClient (ctx context.Context, token, id string) (*Client, error) {
    resp := &getClientResp{}
    
    errObj, err := this.send (ctx, http.MethodGet, token, "Client/get/" + id + "/", nil, resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj.Err() } // something else bad

    if resp.Flag != true {
        return nil, errors.Errorf("response flag was not true : %+v", resp)
    }

    return &resp.Data, nil // we're good
}
