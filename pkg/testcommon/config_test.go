package testcommon

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Assert that it panics if it does not find a config.ini file
	assertPanic(t)

	// Setting the CONFIG_DIR and testing again (supposed to work)
	os.Setenv("CONFIG_DIR", "../..")
	c := LoadConfig("smoke/interioredge")
	if c == nil {
		t.Errorf("Config was not loaded when CONFIG_DIR was set")
	}

	// Validating properties can be loaded and overridden
	if c.GetEnvPropertyBool(PropertyDebug, true) == true {
		t.Errorf("DEBUG property is set to false by default, but true was returned.")
	}

	// Override its value and observe if it works
	os.Setenv("DEBUG", "true")
	if c.GetEnvPropertyBool(PropertyDebug, false) == false {
		t.Errorf("DEBUG property was overridden and set to true, but false was returned.")
	}

	// Validate if the section specific variable is returned
	v, err := c.GetEnvPropertyInt("NUMBER_CLIENTS", 5)
	if err != nil {
		t.Errorf("NUMBER_CLIENTS should return a valid int. Got: %v", err)
	}
	if v != 1 {
		t.Errorf("NUMBER_CLIENTS should be 1 by default, got: %v", v)
	}
	// Overwriting the value of NUMBER_CLIENTS (using string)
	os.Setenv("NUMBER_CLIENTS", "INVALID")
	v, err = c.GetEnvPropertyInt("NUMBER_CLIENTS", 5)
	if err == nil {
		t.Errorf("NUMBER_CLIENTS expected to be invalid. Got: %v", v)
	}

	// Reading invalid key, supposed to return the default value
	dflt := c.GetEnvProperty("INVALID_KEY", "default value")
	if dflt != "default value" {
		t.Errorf("default value expected, got: %v", dflt)
	}

}

func assertPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("LoadConfig() function was expected to panic")
		}
	}()
	LoadConfig("")
}
