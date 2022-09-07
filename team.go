/** ****************************************************************************************************************** **
	Calls related to team members

    
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

type Member struct {
    Id, Name, Role, Email string 
    Active, FieldTech bool 
    ServiceAreas, Skills []string 
}

type memberResponse []*Member
type teamResponse []memberResponse

// takes the jobs out of whatever this parent object is for
func (this teamResponse) toMembers () (ret []*Member) {
    for _, t := range this {
        for _, m := range t {
            ret = append (ret, m)
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

// returns all jobs that match our conditions
func (this *Workiz) ListTeam (ctx context.Context, token string) ([]*Member, error) {
    var resp teamResponse
    
    errObj, err := this.send (ctx, http.MethodGet, token, "team/all/", nil, &resp)
    if err != nil { return nil, errors.WithStack(err) } // bail
    if errObj != nil { return nil, errObj } // something else bad

    return resp.toMembers(), nil // we're good
}