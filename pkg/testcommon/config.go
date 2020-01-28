package testcommon

import (
	"fmt"
	"github.com/go-ini/ini"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	iniFile *ini.File
	Section string
}

func LoadConfig(section string) *Config {
	// If CONFIG_DIR set, use it or use PWD
	configDir := os.Getenv(ConfigDirEnvVar)
	if configDir == "" {
		configDir = os.Getenv("PWD")
	}

	// If no CONFIG_DIR defined, use current directory (empty configDir)
	cfgFileName := configDir + string(os.PathSeparator) + ConfigFile
	f, err := ini.Load(cfgFileName)
	if err != nil {
		panic(fmt.Errorf("unable to read main configuration file for qdr-shipshape: %s - error: %v", cfgFileName, err))
	}

	c := &Config{
		iniFile: f,
		Section: section,
	}
	return c
}

func (c *Config) GetEnvProperty(key string, defaultValue string) string {

	if key == "" {
		return defaultValue
	}

	// First try reading from environment
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	// Next read from specific Section
	return c.GetProperty(c.Section, key, c.GetProperty(DefaultSection, key, defaultValue))

}

func (c *Config) GetEnvPropertyBool(key string, defaultValue bool) bool {
	v := c.GetEnvProperty(key, DefaultSection)
	if v == "" {
		return defaultValue
	}

	if v == "1" || strings.ToLower(v) == "true" {
		return true
	}

	return false
}

func (c *Config) GetEnvPropertyInt(key string, defaultValue int) (int, error) {
	v := c.GetEnvProperty(key, DefaultSection)

	if v == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(v)
}

func (c *Config) GetProperty(section, key, defaultValue string) string {
	// Next read from specific Section
	value := c.iniFile.Section(section).Key(key).String()
	if value != "" {
		return value
	}

	return defaultValue
}

