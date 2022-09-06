package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Path for directory that stores config files and logs. Defaults to /etc/spirit-box/.
var SPIRIT_PATH string

// Path to write debug logs to. Defaults to /dev/null.
var DEBUG_FILE string

// Defaults to false.
var TUI_FANCY bool

// Permission for user to view expanded info on systemd units.
var SYSTEMD_ACCESS bool

// Run spirit-box or exit early.
var ENABLED bool

var GENERAL_CONFIG_PATH = "config.json"
var NETWORK_CONFIG_PATH = "network.json"
var SCRIPT_SPEC_PATH = "script_specs.json"
var UNIT_SPEC_PATH = "unit_specs.json"
var LOG_PATH = "logs/"

// Sets up globals, should be called after flags have been parsed.
func InitPaths() {
	GENERAL_CONFIG_PATH = SPIRIT_PATH + GENERAL_CONFIG_PATH
	NETWORK_CONFIG_PATH = SPIRIT_PATH + NETWORK_CONFIG_PATH
	SCRIPT_SPEC_PATH = SPIRIT_PATH + SCRIPT_SPEC_PATH
	UNIT_SPEC_PATH = SPIRIT_PATH + UNIT_SPEC_PATH
	LOG_PATH = SPIRIT_PATH + LOG_PATH
}

type GeneralConfig struct {
	SystemdAccess bool `json:"systemdAccess"`
	Enabled       bool `json:"enabled"`
}

func LoadGeneralConfig() {
	// TODO: allow for overrides from a location in persistent memory.
	type ParseObj struct {
		Config GeneralConfig `json:"config"`
	}

	temp := ParseObj{}

	bytes, err := os.ReadFile(GENERAL_CONFIG_PATH)
	if err != nil {
		log.Fatal(fmt.Errorf("Loading general config: %s", err.Error()))
	}

	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		log.Fatal(fmt.Errorf("Loading general config: %s", err.Error()))
	}

	SYSTEMD_ACCESS = temp.Config.SystemdAccess
	ENABLED = temp.Config.Enabled
}
