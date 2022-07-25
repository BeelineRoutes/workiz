/** ****************************************************************************************************************** **
	The actual sending and receiving stuff
	Reused for most of the calls to Workiz
	
** ****************************************************************************************************************** **/

package workiz 

import (
    "github.com/pkg/errors"

    "fmt"
    "net/http"
    "context"
    "encoding/json"
    "io/ioutil"
    "bytes"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// handles making the request and reading the results from it 
// if there's an error the Error object will be set, otherwise it will be nil
func (this *Workiz) finish (req *http.Request, out interface{}) (*Error, error) {
	resp, err := http.DefaultClient.Do (req)
	
	if err != nil { return nil, errors.WithStack (err) }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll (resp.Body)

    if resp.StatusCode > 399 { 
		errObj := &Error{
			StatusCode: resp.StatusCode,
			Msg: string(body),
		}
        
        return errObj, errors.Wrap (err, string(body))
    }
	
	if out != nil { err = errors.WithStack (json.Unmarshal (body, out)) }
	
	return nil, err // we're good
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

func (this *Workiz) send (ctx context.Context, requestType, token, link string, in, out interface{}) (*Error, error) {
	var jstr []byte 
	var err error 

	header := make(map[string]string)

	if in != nil {
		jstr, err = json.Marshal (in)
		if err != nil { return nil, errors.WithStack (err) }

		header["Content-Type"] = "application/json; charset=utf-8"
	}
	
	req, err := http.NewRequestWithContext (ctx, requestType, fmt.Sprintf ("%s/%s/%s", apiURL, token, link), bytes.NewBuffer(jstr))
	if err != nil { return nil, errors.Wrap (err, link) }

	for key, val := range header { req.Header.Set (key, val) }
	errObj, err := this.finish (req, out)
	
	return errObj, errors.Wrapf (err, " %s : %s", link, string(jstr))
}
