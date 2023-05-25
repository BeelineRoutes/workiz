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
	"strings"
	"time"
	"math"
)

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE ---------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// handles making the request and reading the results from it 
// if there's an error the Error object will be set, otherwise it will be nil
func (this *Workiz) finish (req *http.Request, out interface{}) error {
	resp, err := http.DefaultClient.Do (req)
	
	if err != nil { return errors.WithStack (err) }
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll (resp.Body)

    if resp.StatusCode > 399 { 
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			// special error 
			return errors.Wrapf (ErrAuthExpired, "Unauthorized : %d : %s", resp.StatusCode, string(body))

		case http.StatusTooManyRequests:
			return errors.Wrapf (ErrQuota, "Quota : %d : %s", resp.StatusCode, string(body))
		}
		// just a default
		err = errors.Wrapf (ErrUnexpected, "Workiz Error : %d : %s", resp.StatusCode, string(body))

		// see if we can figure out the error
		errResp := &apiResp{}
		jErr := errors.WithStack (json.Unmarshal (body, errResp))
		if jErr == nil {
			// i want to try to "handle" some of these errors here that aren't actually errors
			if errResp.Code == 400 {
				if strings.Contains(errResp.Details.Error, "User is already assigned") {
					return nil
				}
				if strings.Contains(errResp.Details.Error, "User is not assigned") {
					return nil
				}
			}

			// we don't know what to do with this error
			err = errors.Wrapf (err, "%s : %s", errResp.Details.Error, errResp.Msg)
		} else {
			// different error object than expected
			err = errors.Wrapf (err, "unmarshal : %s", jErr.Error())
		}
        
        return err
    }
	
	if out != nil { 
		err = errors.WithStack (json.Unmarshal (body, out))
		if err != nil {
			err = errors.Wrap (err, string(body)) // if it didn't unmarshal, include the body so we know what it did look like
		}
	}
	
	return err // we're good
}

  //-----------------------------------------------------------------------------------------------------------------------//
 //----- FUNCTIONS -------------------------------------------------------------------------------------------------------//
//-----------------------------------------------------------------------------------------------------------------------//

// this recurses
// retries itself on a 429 - ErrQuota
func (this *Workiz) send (ctx context.Context, retries int, requestType, token, link string, in, out interface{}) error {
	if ctx.Err() != nil { return ctx.Err() } // bail on a context timeout

	var jstr []byte 
	var err error 

	header := make(map[string]string)

	if in != nil {
		jstr, err = json.Marshal (in)
		if err != nil { return errors.WithStack (err) }

		header["Content-Type"] = "application/json; charset=utf-8"
	}
	
	req, err := http.NewRequestWithContext (ctx, requestType, fmt.Sprintf ("%s/%s/%s", apiURL, token, link), bytes.NewBuffer(jstr))
	if err != nil { return errors.Wrap (err, link) }

	for key, val := range header { req.Header.Set (key, val) }
	err = this.finish (req, out)

	switch errors.Cause(err) {
	case ErrQuota:
		if retries < 6 { // 6 gives 1 + 3 + 7 + 15 + 31 + 63 seconds wait
			time.Sleep (time.Second * time.Duration(math.Pow(2, float64(retries)))) // exp timeout for sleeping
			return this.send (ctx, retries +1, requestType, token, link, in, out)
		}
	}
	
	return errors.Wrapf (err, " %s : %s", link, string(jstr))
}
