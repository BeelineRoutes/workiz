
package workiz 

import (
	"github.com/stretchr/testify/assert"

	"testing"
	"encoding/json"
)

func getRealConfig (t *testing.T) *Config {
	config, err := parseConfig("./config.json") // this config works and isnt' included in the repo
	if err != nil { t.Fatal(err) } // should have worked
	if config.Valid() == false {
		t.Fatal("config file 'config.json' doesn't appear valid")
	}
	return config
}

func TestConfig (t *testing.T) {
	config, err := parseConfig("./example_config.json")
	if err != nil { t.Fatal(err) } // should have worked

	assert.Equal (t, false, config.Valid())
	assert.Equal(t, "api_123...", config.Token)
	assert.Equal(t, "sec_123...", config.Secret)
}

func TestApiRespError1 (t *testing.T) {
	one := &apiResp{}

	err := json.Unmarshal([]byte(`{"error":true,"code":429,"msg":"Account reached Api Quotas - Please try again later or talk to our support to increase quota","details":[]}`), one)
	if err != nil { t.Fatal(err) }

	assert.Equal (t, true, one.Error)
	assert.Equal (t, 429, one.Code)

}

func TestApiRespError2 (t *testing.T) {
	one := &apiResp{}

	err := json.Unmarshal([]byte(`{"error":true,"code":400,"msg":"Validation rule exception","details":{"error":"Cannot assign Alissa Thomas to job XZDO9T, User is already assigned."}}`), one)
	if err != nil { t.Fatal(err) }

	assert.Equal (t, true, one.Error)
	assert.Equal (t, 68, len(one.Details.Error))

}
