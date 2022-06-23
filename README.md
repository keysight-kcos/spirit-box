# data-driven-boot-up-ui
Data-driven Web/Terminal UI for a Linux System Boot-Up

### TODO
- [ ] The tracking of services sometimes misses updates; change the implementation to have consistent tracking.
- [ ] Investigate ways to remove the need to switch between TTYs.

## 06/23
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