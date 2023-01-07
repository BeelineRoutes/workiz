
package workiz 

import (
	"github.com/stretchr/testify/assert"

	"testing"
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

