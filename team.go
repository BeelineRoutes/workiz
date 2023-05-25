/** ****************************************************************************************************************** **
	Calls related to team members

    
** ****************************************************************************************************************** **/

package workiz 

import (
    //"fmt"
    "net/http"
    "context"
    "strings"
    
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

type Members []*Member 

func (this *Members) Push (in *Member) {
    *this = append(*this, in)
}

// loops through the list and returns the id based on the name
func (this Members) FindId (nm string) string {
    for _, m := range this {
        if strings.EqualFold(m.Name, nm) { return m.Id }
    }

    return "" // didn't find them
}

// loops thruogh and finds a name from the id
func (this Members) FindName (id string) string {
    for _, m := range this {
        if strings.EqualFold (m.Id, id) { return m.Name }
    }

    return "" // didn't find them
}

type teamResponse struct {
    Data []*Member
}

// takes the jobs out of whatever this parent object is for
func (this teamResponse) toMembers () (ret Members) {
    for _, m := range this.Data {
        if m.Active == false { continue }
        if m.FieldTech == false { continue }

        // they're good to get jobs
        ret.Push(m)
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
func (this *Workiz) ListTeam (ctx context.Context, token string) (Members, error) {
    var resp teamResponse
    
    err := this.send (ctx, 0, http.MethodGet, token, "team/all/", nil, &resp)
    if err != nil { return nil, err } // bail
    
    return resp.toMembers(), nil // we're good
}
