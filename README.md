# data-driven-boot-up-ui
Data-driven Web/Terminal UI for a Linux System Boot-Up

### Links
Terminal cursor movement & overwriting lines: https://unix.stackexchange.com/questions/43075/how-to-change-the-contents-of-a-line-on-the-terminal-as-opposed-to-writing-a-new

### TODO
- [ ] Investigate ways to remove the need to switch between TTYs.
- [ ] Create a log file with timestamps.

## 06/23
- [x] The tracking of services sometimes misses updates; change the implementation to have consistent tracking.

The states of the units is now tracked on a set interval. The program
appears to catch all updates now.

Also, I added some terminal escape codes to overwrite updates on the screen
(as opposed appending each update as a list) as well as added output that
displays the amount of time that has passed. A sequential list of the
status of the units could be written to a log file with timestamps.

## 06/22
As of now, the printSystemdInfo program will track the services in the whitelist:
```
echo_server.service:running
printSpam.service:dead
polkit.service:running
```
where the format is \<unit name\>:\<substate\>. 

When the service has the same substate that is
specified in the whitelist, it is considered "ready".

Example output:
```
Units to be watched:
echo_server.service, ready when substate=running
printSpam.service, ready when substate=dead
polkit.service, ready when substate=running

Timeout = 120s

Initial states:
echo_server.service: loaded active running
printSpam.service: loaded inactive dead
polkit.service: loaded active running

Waiting for unit updates...
echo_server.service: ready=true
printSpam.service: ready=true
polkit.service: ready=true

3 units are ready.
```

The systemd unit file for launching this program:
```
[Unit]
Description=Print systemd info to tty.
After=dbus.service
Before=getty@tty2.service
StartLimitIntervalSec=0

[Service]
Type=oneshot
ExecStart=/<path_to>/printSystemdInfo
StandardOutput=tty
TTYPath=/dev/tty2

[Install]
WantedBy=multi-user.target
```

This will launch the binary for the program and output to tty2.

In addition, we must add the following file to ensure that the output printed to tty2 
is not cleared after the login prompt pops up:

`/etc/systemd/system/getty@tty2.service.d/override.conf`

Contents of override.conf:
```
[Service]
TTYVTDisallocate=no
```