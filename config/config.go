package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"spirit-box/device"
	"spirit-box/logging"
	"spirit-box/scripts"
	"spirit-box/services"
)

// Path for directory that stores config files and logs. Defaults to /etc/spirit-box/.
var SPIRIT_PATH string

// Path to write debug logs to. Defaults to /dev/null.
var DEBUG_FILE string

// Defaults to false.
var TUI_FANCY bool

// Permission for user to view expanded info on systemd units.
var SYSTEMD_ACCESS bool

// Message to display when system is ready.
var BANNER_MESSAGE string

// Run spirit-box or exit early.
var ENABLED bool

var CONFIG_PATH = "config.json"
var NETWORK_CONFIG_PATH = "network.json"
var SCRIPT_SPEC_PATH = "script_specs.json"
var UNIT_SPEC_PATH = "unit_specs.json"
var LOG_PATH = "logs/"

// Sets up globals, should be called after flags have been parsed.
func initPaths() {
	CONFIG_PATH = SPIRIT_PATH + CONFIG_PATH
	NETWORK_CONFIG_PATH = SPIRIT_PATH + NETWORK_CONFIG_PATH
	SCRIPT_SPEC_PATH = SPIRIT_PATH + SCRIPT_SPEC_PATH
	UNIT_SPEC_PATH = SPIRIT_PATH + UNIT_SPEC_PATH
	LOG_PATH = SPIRIT_PATH + LOG_PATH
}

type ParseObj struct {
	ServerPort     string               `json:"serverPort"`
	HostPort       string               `json:"hostPort"`
	TempPort       string               `json:"tempPort"`
	Nic            string               `json:"nic"`
	SystemdAccess  string               `json:"systemdAccess"`
	BannerMessage  string               `json:"bannerMessage"`
	Enabled        string               `json:"enabled"`
	UnitSpecArr    []services.UnitSpec  `json:"unitSpecs"`
	ScriptSpecArr  []scripts.ScriptSpec `json:"scriptSpecs"`
	ConfigOverride string               `json:"configOverride"`
}

func LoadConfig() {
	initPaths()
	configObj := ParseObj{}

	bytes, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		log.Fatal(fmt.Errorf("Loading config from %s: %s", CONFIG_PATH, err.Error()))
	}

	err = json.Unmarshal(bytes, &configObj)
	if err != nil {
		log.Fatal(fmt.Errorf("Loading config from %s: %s", CONFIG_PATH, err.Error()))
	}
	log.Printf("Successfully loaded config from %s.", CONFIG_PATH)

	loadConfigRecursive(&configObj, configObj.ConfigOverride)
	log.Printf("Successfully loaded config: %+v\n", configObj)

	// not boolean for purpose of conditional overrides
	if configObj.Enabled == "true" {
		ENABLED = true
	}
	if configObj.SystemdAccess == "true" {
		SYSTEMD_ACCESS = true
	}
	BANNER_MESSAGE = configObj.BannerMessage

	logging.LOG_PATH = LOG_PATH

	device.SERVER_PORT = configObj.ServerPort
	device.HOST_PORT = configObj.HostPort
	device.TEMP_PORT = configObj.TempPort
	device.NIC = configObj.Nic

	scripts.SCRIPT_SPECS = configObj.ScriptSpecArr
	services.UNIT_SPECS = configObj.UnitSpecArr
}

func loadConfigRecursive(configObj *ParseObj, configPath string) {
	fileInfo, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		log.Printf("Loading config from %s: Path does not exist.", configPath)
		return
	}
	if fileInfo.IsDir() {
		log.Printf("Loading config from %s: Is a directory.", configPath)
		return
	}

	temp := ParseObj{}
	bytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Print(fmt.Errorf("Loading config from %s: %s", configPath, err.Error()))
		return
	}

	err = json.Unmarshal(bytes, &temp)
	if err != nil {
		log.Print(fmt.Errorf("Loading config from %s: %s", configPath, err.Error()))
		return
	}
	log.Printf("Successfully loaded config from %s.", configPath)

	joinConfigs(configObj, &temp)
	loadConfigRecursive(configObj, temp.ConfigOverride)
}

func joinConfigs(configObj *ParseObj, overrides *ParseObj) {
	// empty string if field was ommitted.
	if overrides.ServerPort != "" {
		configObj.ServerPort = overrides.ServerPort
	}
	if overrides.HostPort != "" {
		configObj.HostPort = overrides.HostPort
	}
	if overrides.TempPort != "" {
		configObj.TempPort = overrides.TempPort
	}
	if overrides.Nic != "" {
		configObj.Nic = overrides.Nic
	}
	if overrides.SystemdAccess != "" {
		configObj.SystemdAccess = overrides.SystemdAccess
	}
	if overrides.BannerMessage != "" {
		configObj.BannerMessage = overrides.BannerMessage
	}
	if overrides.Enabled != "" {
		configObj.Enabled = overrides.Enabled
	}

	if len(overrides.UnitSpecArr) > 0 {
		for _, spec := range overrides.UnitSpecArr {
			duplicate := false
			sig := spec.ToString()

			for _, curSpec := range configObj.UnitSpecArr {
				if curSpec.ToString() == sig {
					duplicate = true
					break
				}
			}
			if !duplicate {
				configObj.UnitSpecArr = append(configObj.UnitSpecArr, spec)
			}
		}
	}

	if len(overrides.ScriptSpecArr) > 0 {
		for _, spec := range overrides.ScriptSpecArr {
			duplicate := false
			sig := fmt.Sprintf("%v", spec)

			for _, curSpec := range configObj.ScriptSpecArr {
				if fmt.Sprintf("%v", curSpec) == sig {
					duplicate = true
					break
				}
			}
			if !duplicate {
				configObj.ScriptSpecArr = append(configObj.ScriptSpecArr, spec)
			}
		}
	}
}
