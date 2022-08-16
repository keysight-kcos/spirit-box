# spirit-box
insert logo here

## Intro
For machines with long boot processes, it is impossible to know exactly what the system is doing. Spirit-box is the solution. Spirit-box is a data-driven web and terminal UI for Linux systems. It provides real-time updates for visibility into the system during early stages of boot. 

Currently, it is possible to get live updates on systemd units. Users are also able to use their own scripts to give more precise insight into the processes that are being started. All of this is data-driven and configurable by the user.

![mainScreen](https://user-images.githubusercontent.com/56091505/179314635-4c7cb978-9708-4d54-8860-7944feb22b97.gif)


## Systemd Units
Users can define what units they would like to monitor in the whitelist file. The service will be considered ready by spirit-box once it achieves the specified substate. Each service is seperated by a newline. The format is:

    unitname:substate

For example,

    NetworkManager.service:running

will have spirit-box monitor the network manager service.

## Scripts
Users can specify scripts for spirit-box to run on boot. These scripts are typically used to monitor boot-up processes in greater detail than systemd can provide. However, they do not have to be monitor scripts. The list of scripts are defined in script_specs.json. The file is an array of objects. The format of each script is:

+ "cmd": "<path to script\>",

+ "args": ["<argument 1\>", "<argument 2\>" ...],

+ "priority": <priority level\>,

+ "retryTimeout": <time between retries of the script\>,

+ "totalWaitTime": <time until considering the script a failure\>

Scripts can be sorted by priority. All scripts in a lower priority group must either complete successfully or time out before spirit-box begins to execute scripts of the next higher priority group. This allows scripts to be run after their dependancies have been successfully finished. Scripts with the same priority run concurrently.

retryTimeout is the amount of time (in milliseconds) spirit-box will wait before attempting to rerun the script. Spirit-box will only attempt to rerun scripts if they time out from this timer, or if they return as unsuccessful. Unsuccessful returns are defined in the script by the user.

totalWaitTime is the amount of time (in milliseconds) spirit-box will wait for a script to return successfully before declaring the script as a failure. Spirit-box will stop rerunning this script. This time includes all the reruns of a script.

## Logging


Spirit-box creates comprehensive logs detailing the systemd services and scripts specified in the configurations. All logs are organized in the JSON format. Each log event has five fields:

+ startTime - the time the event was observed by spirit-box
+ endTime - the time the event concluded
+ description - a short information excerpt of the event
+ objectType - the type of event that occured
+ object - contains additional data related to the event

There are multiple objectTypes that represent different sorts of events.

+ Message - details critical information about spirit-box such as when spirit-box starts and its dependancies are up
+ SystemD unit state change - describes state and substate changes in a systemd unit. Substate data is contained in the object.
+ Script event - describes script executions. The object contains data from every run of the script, if the script was rerun due to failure. It contains data such as the script's command path, arguments, priority group, timeouts, and success status.

## Terminal User Interface

The spirit-box terminal user interface is displayed on boot. The main screen displays the status of all systemd units as well as all scripts. It displays an IP and port to the webpage hosting the graphical user interface. The main screen has live updates whenever a new event is observed by spirit-box. The user is able to select whether they would like to view the systemd screen or the scripts screen.

The systemd screen has an overview of all whitelisted services. It displays their substates and ready status. The user is able to add services to watch at run time. A list of properties and their values are accessible when the user selects the service.

![Screenshot 2022-07-15 153240](https://user-images.githubusercontent.com/56091505/179320455-3766f4fc-3fbf-487b-9ab0-58fc4257a4e8.png)

The scripts screen has an overview of all scripts specified in the configuration files. Scripts are organized by priority group. Selecting a priority group allows the user to view information about the scripts within that group.

![Screenshot 2022-07-18 154508](https://user-images.githubusercontent.com/56091505/179629671-bdba3352-9e1c-4ff6-bc90-871bbaa200f7.png)

## Graphical User Interface

Spirit-box's web page can be seen on localhost both port 80 and port 8080. Once spirit-box exits, port 80 will be open to any other services that uses it automatically. Reload the page to see what's being served on port 80. The web UI updates the progress of systemd services and scripts automatically. Shutting down spirit-box from the web UI will exit the terminal UI.

![Screenshot 2022-07-18 161909](https://user-images.githubusercontent.com/56091505/179632771-941def88-4ffe-4be2-86fd-11853c777368.png)

The scripts tab has an overview of all scripts specified in the configuration files. Scripts are organized by priority group. The user is able to view information about the scripts within that group. Selecting a script shows information about each run.

The systemd tab has an overview of all whitelisted services. It displays their substates and ready status. A list of properties and their values are accessible when the user selects the service.

## Tutorial

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

## 08/16
Here are our current todos listed in approximate order of importance:

- [x] Graceful handoff between spirit-box web UI and host machine's web UI
- [x] Option for a sequential view of all checks -> TUI lite
- [x] Timestamps and PID for each individual script run -> this is now included in logs
- [x] Exact times for systemd timestamps rather than observed times -> events are now ns precision
- [x] Build a demo environment that demonstrates how this tool can be used in the wild, use KCOS usecases as a base
- [x] Run the program on different types of devices -> tested on KCOS, CentOS, Ubuntu
- [ ] Documentation to make it easy for anyone to jump into the project and add things
- [ ] Documentation for how the program is used
- [x] Streamlined installations onto host machines with minimal tinkering -> easy to install deb packages
- [ ] Log visualization -> timeline, graphs, etc. thru grafana
- [ ] Productization -> Make it easy for others to decide if they have a need for this tool
- [ ] Option to disable the ability to add more systemd units during runtime
- [ ] Interactivity with script execution
- [ ] Combine config files into a single file if that would provide any benefit
- [ ] Possible use of eBPF
- [ ] Performance and memory profiling

