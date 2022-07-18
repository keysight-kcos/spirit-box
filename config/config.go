package config

// Path for directory that stores config files and logs.
var SPIRIT_PATH string // defaults to /etc/spirit-box/

var NETWORK_CONFIG_PATH = "network.json"
var SCRIPT_SPEC_PATH = "script_specs.json"
var WHITELIST_PATH = "whitelist"
var LOG_PATH = "logs/"

// Sets up globals, should be called after flags have been parsed.
func InitPaths() {
	NETWORK_CONFIG_PATH = SPIRIT_PATH + NETWORK_CONFIG_PATH
	SCRIPT_SPEC_PATH = SPIRIT_PATH + SCRIPT_SPEC_PATH
	WHITELIST_PATH = SPIRIT_PATH + WHITELIST_PATH
	LOG_PATH = SPIRIT_PATH + LOG_PATH
}
