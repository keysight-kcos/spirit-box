# spirit-box

## 07/14

Program behavior:

While the program is running:
	- spirit-box's web page can be seen on both port 80 and port 8080
	- Once there's anything being served on port 80 the iptables rules are undone. Reload page to see what's being served on port 80.

The iptables commands in the program need root permission so build main.go first and run with sudo.

## 07/12
Presented progress demo today. Here are our current todos:

- [ ] Interactivity with script execution
- [ ] Timestamps and PID for each individual script run
- [ ] Exact times for systemd timestamps rather than observed times. The extra precision would be great but the method for getting exact timestamps needs to be explored.
- [ ] Graceful handoff between spirit-box web UI and host machine's web UI.
- [ ] Log visualization -> timeline, graphs, etc.
- [ ] Possible use of eBPF
- [ ] Streamlined installations onto host machines with minimal tinkering
- [ ] Combine config files into a single file if that would provide any benefit
- [ ] Documentation to make it easy for anyone to jump into the project and add things
- [ ] Documentation for how the program is used
- [ ] Option for a sequential view of all checks
- [ ] Option to disable the ability to add more systemd units during runtime
- [ ] Run the program on different types of devices
- [ ] Performance and memory profiling
- [ ] Productization -> Make it easy for others to decide if they have a need for this tool
- [ ] Build a demo environment that demonstrates how this tool can be used in the wild, use KCOS usecases as a base

## 07/07
Merging a bunch of stuff into main.

- Terminal interface.
- React frontend for a web interface.
- Improved logging.
- Scripts functionality will be added soon.

___
## 06/27
We are now looking into using charm's bubbletea package for implementing TUI functionality.
- [x] Basic networking information (IPs for each NIC)
- [x] Implement logging of systemd tracking info.
- [x] Move systemd monitoring functions to their own subpackage.

## 06/24
___

As of now, our two main goals are to add logging and the ability for the program to execute arbitrary scripts/programs (command paths read from a file).

After these are done, we want to use goroutines and channels to simultaneously output the systemd info and the script info (probably just output and exit codes for now).

The next step after that would possibly be to expand the script execution functionality by defining some standardized interface for the scripts that are run so that users have access to more fine-grained information in our program's output.

- Created a .deb package. Was able to install and run on Lorenzo's machine. Issue with the disableSystemdLogging.service file, but I believe it is fixed now. Will confirm on Monday.
- Preliminary logging subpackage in place. Has yet to be implemented with current systemd tracking functions.
___
## 06/23

___
## - [x] Investigate ways to remove the need to switch between TTYs.

When using a serial terminal, the user cannot switch between TTYs.
Therefore, it is important to make sure our program can print to a single
terminal and delay the login prompt from popping up until the program is 
complete.

First, I disabled the startup of X at boot using information from the
following SO post: https://askubuntu.com/questions/16371/how-do-i-disable-x-at-boot-time-so-that-the-system-boots-in-text-mode

These are the exact steps I took:
1. In `/etc/default/grub` I changed 

    `GRUB_CMDLINE_LINUX_DEFAULT="quiet splash video=hyperv_fb:1920x1080"`

    to

    `GRUB_CMDLINE_LINUX_DEFAULT="text video=hyperv_fb:1920x1080"`

    The "video=*" snippet is something I had added earlier to change the screen size of the VM. By default it would not be in the grub file.

2. I ran the command `sudo update-grub`.
3. I ran the command `systemctl get-default` and made note of the output (graphical.target on my machine) in case I want to reverse these changes in the future.
4. I ran the command `systemctl set-default multi-user.target`.

After this, the machine would boot into a tty on startup. When I wanted to start X, I ran the command `systemctl start lightdm`. This is not the only
way to start X and display managers differ between distributions.

Next, I created a unit file `/etc/systemd/system/disableSystemdLogging.service` to disable systemd logging on boot, using a unit file within kcos-ghost as a reference: 
```
[Unit]
Description=Disable systemd console logging.
StartLimitIntervalSec=0
After=systemd
Before=time-set.target

[Service]
Type=oneshot
RemainAfterExit=yes
# send signal to disable console logging
ExecStart=kill -s SIGRTMIN+21 1
# send signal to enable console logging
# ExecStop=kill -s SIGRTMIN+20 1

[Install]
WantedBy=multi-user.target
```

Figuring out the right place to have this service run is a work in progress, but after the service completes it does silence further logging from systemd. For now, I just need to prevent systemd's logging from interfering with the output of the `printSystemdInfo` service.

I edited the `printSystemdInfo.service` to ensure that systemd logging is disabled before the service runs. I also changed the TTYPath to the bootup tty on this VM. This may be different from machine to machine.
```
[Unit]
Description=Print systemd info to tty.
After=dbus.service disableSystemdLogging.service
StartLimitIntervalSec=0

[Service]
Type=oneshot
ExecStart=/home/severian/data-driven-boot-up-ui/printSystemdInfo
StandardOutput=tty
TTYPath=/dev/tty1

[Install]
WantedBy=multi-user.target
```

Finally, I edited `/etc/systemd/system/getty@tty1.service.d/override.conf` to make sure that the `printSystemdInfo` service has completed before the login prompt is presented:
```
[Unit]
After=printSystemdInfo.service

[Service]
TTYVTDisallocate=no

```
___

## - [x] The tracking of services sometimes misses updates; change the implementation to have consistent tracking.

The states of the units is now tracked on a set interval. The program
appears to catch all updates now.

Also, I added some terminal escape codes to overwrite updates on the screen
(as opposed appending each update in a sequential list format) 
In addition, I added some output that displays the amount of time that has passed. A sequential list of the
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
