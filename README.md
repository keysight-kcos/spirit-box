# spirit-box

spirit-box provides a real-time visibility window into Linux systems during boot-up. 

![mainScreen](https://user-images.githubusercontent.com/56091505/179314635-4c7cb978-9708-4d54-8860-7944feb22b97.gif)

This visibility is achieved through live updates on the state of systemd units as well as through user-specified custom status-checking scripts.

spirit-box aims to be data-driven and highly configurable, allowing for users to adjust the tool to match the visibility needs of their target system.

## Config

spirit-box's primary configuration file is located in `/etc/spirit-box/` by default. The location of the `spirit-box` directory can be set
when running spirit-box with the flag `-p <path to directory>`.

- `serverPort`: The port that spirit-box's server uses.
- `hostPort`: The port that the host machine uses for its web UI (if it has one).
- `tempPort`: The port to which the host machine's default web server is rerouted while spirit-box is running.
- `nic`: The nic on which to set iptables rules and gather IP addresses.
- `systemdAccess`: Control's the user's access to the full readouts of systemd units.
- `bannerMessage`: A message to display after spirit-box recognizes that the host system is ready.
- `enabled`: Determines whether spirit-box runs normally or exits early.
- `configOverride`: A path to an override config file. Fields that are set in an override file will override the fields set in previous config files, except for
the `unitSpecs` and `scriptSpecs` fields, which will append specifications instead. These can be chained indefinitely, but there are currently no checks for loops. 
- `unitSpecs`: 
    - `name`: The name of the systemd unit to be tracked.
    - `desc`: An alias used for the unit when displayed in the spirit-box UIs.
    - `substateDesired`: The state at which the unit is considered ready.
 - `scriptSpecs`:
    - `cmd`: The path to the script's executable.
    - `args`: Arguments passed to the script.
    - `desc`: An alias used for the script when displayed in the spirit-box UIs.
    - `priority`: A positive number specifying the order in which scripts are run. Scripts with lower priority numbers are run first. Scripts with the same priority number are run concurrently.
    - `retryTimeout`: The amount of time to wait before rerunning a script if it has failed.
    - `totalWaitTime`: The max amount of time to wait for a script to return a success, including reruns.

Example `config.json` file. 
```
{
	"serverPort": "8080",
	"hostPort": "80",
	"tempPort": "8081",
	"nic": "eth0",
	"systemdAccess": "true",
	"bannerMessage": "Hack the planet.",
	"enabled": "true",
	"configOverride": "/nonexistent/path/override1.json",
	"unitSpecs": [
		{
			"name": "polkit.service",
			"desc": "polkit",
			"subStateDesired": "running"
		},
		{
			"name": "NetworkManager.service",
			"desc": "network manager",
			"subStateDesired": "running"
		},
		{
			"name": "cron.service",
			"desc": "cron",
			"subStateDesired": "running"
		},
		{
			"name": "docker.service",
			"desc": "docker",
			"subStateDesired": "running"
		},
		{
			"name": "printSpam.service",
			"desc": "printSpam",
			"subStateDesired": "dead"
		}
	],
	"scriptSpecs": [
		{
			"cmd": "/usr/bin/dummyScript",
			"args": ["-wait=500", "-prob=30"],
			"desc": "dummy 30",
			"priority": 1,
			"retryTimeout": 150,
			"totalWaitTime": 3000
		},
		{
			"cmd": "/usr/bin/dummyScript",
			"args": ["-wait=500"],
			"desc": "dummy 50",
			"priority": 1,
			"retryTimeout": 200,
			"totalWaitTime": 3000
		},
		{
			"cmd": "/usr/bin/dummyScript",
			"args": ["-wait=1500", "-prob=60"],
			"desc": "dummy 60",
			"priority": 2,
			"retryTimeout": 150,
			"totalWaitTime": 3900
		},
		{
			"cmd": "/usr/bin/dummyScript2",
			"args": ["-wait=1500", "-prob=60"],
			"desc": "dummy 60 2",
			"priority": 2,
			"retryTimeout": 150,
			"totalWaitTime": 3900
		},
		{
			"cmd": "/usr/bin/dummyScript",
			"args": ["-wait=1500", "-prob=0"],
			"desc": "dummy staller",
			"priority": 2,
			"retryTimeout": 150,
			"totalWaitTime": 3900
		}
	]
}
```

## Script Output Format

The scripts provided to spirit-box are not limited in what they are allowed to do, but their output must follow the following format:
```
{
    "info": "Any arbitrary string",
    "success": <true or false>
}
```
- `info`: This field can be used to capture some state that the script observed if anything more complex than a simple true/false needs to be recorded.
- `success`: If true, the spirit-box will register the check as a success and stop trying to rerun the script. Otherwise, the script will continue to be run within the constraints of the `retryTimeout` and `totalWaitTime` specifications.

## Logging


spirit-box creates comprehensive logs detailing the systemd services and scripts specified in the configurations. All logs are organized in the JSON format. Each log event has five fields:

+ startTime - the time the event was observed by spirit-box
+ endTime - the time the event concluded
+ description - a short information excerpt of the event
+ objectType - the type of event that occured
+ object - contains additional data related to the event

There are multiple objectTypes that represent different sorts of events.

+ Message - details critical information about spirit-box such as when spirit-box starts and its dependancies are up
+ SystemD unit state change - describes state and substate changes in a systemd unit. Substate data is contained in the object.
+ Script event - describes script executions. The object contains data from every run of the script, if the script was rerun due to failure. It contains data such as the script's command path, arguments, priority group, timeouts, and success status.

Log files are stored in the `logs` directory of the spirit-box directory (`/etc/spirit-box/` by default).

## Terminal User Interface

The spirit-box terminal user interface is displayed on boot. The main screen displays the status of all systemd units as well as all scripts. It displays an IP and port to the webpage hosting the graphical user interface. The main screen has live updates whenever a new event is observed by spirit-box. The user is able to select whether they would like to view the systemd screen or the scripts screen.

The systemd screen has an overview of all whitelisted services. It displays their substates and ready status. The user is able to add services to watch at run time. A list of properties and their values are accessible when the user selects the service.

![Screenshot 2022-07-15 153240](https://user-images.githubusercontent.com/56091505/179320455-3766f4fc-3fbf-487b-9ab0-58fc4257a4e8.png)

The scripts screen has an overview of all scripts specified in the configuration files. Scripts are organized by priority group. Selecting a priority group allows the user to view information about the scripts within that group.

![Screenshot 2022-07-18 154508](https://user-images.githubusercontent.com/56091505/179629671-bdba3352-9e1c-4ff6-bc90-871bbaa200f7.png)

## Graphical User Interface

spirit-box's web UI is served on `serverPort` and `hostPort` while it is running. 
Once spirit-box registers that the system is ready, or if it exits early, `hostPort` will be handed back to the host machine's default service. 

![Screenshot 2022-07-18 161909](https://user-images.githubusercontent.com/56091505/179632771-941def88-4ffe-4be2-86fd-11853c777368.png)

The Scripts dashboard displays an overview of the statuses of all scripts specified within the configuration files. Clicking on a script will display the output of each individual run of that script.

The Units dashboard displays the statuses of all systemd units spirit-box is monitoring. If `systemdAccess` is set to "true", clicking on a unit will display all properties about that particular unit.

## Installation

Refer to the README in the `packages` directory for instructions on installing spirit-box via a deb package. 

NOTE: spirit-box is ultimately just a binary that requires a dedicated directory somewhere on the host machine to store an initial config file and store logs. The program does not need to be installed through a deb package. Adjust installation methods according to your usecase.

### Installation methods below are outdated

### Ubuntu/Debian Install
1. download the DEB package

`wget -c https://github.com/keysight-kcos/data-driven-boot-up-ui/releases/download/v1.0.0-alpha/spirit-box_1.0.0_arm64.deb`

2. unpack

`dpkg -i spirit-box_...arm64.deb`

3. enable 

`systemctl enable spirit-box@ttyS0.service`

4. reboot

### Fedora/RHEL Install
1. download the RPM package

`wget -c https://github.com/keysight-kcos/data-driven-boot-up-ui/releases/download/v1.0.0-alpha/spirit-box-1.0.0-2.x86_64.rpm`

2. unpack

`rpm -ivh spirit-box-...x86_64.rpm`

3. enable 

`systemctl enable spirit-box@ttyS0.service`

4. reboot
