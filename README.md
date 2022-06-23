# data-driven-boot-up-ui
Data-driven Web/Terminal UI for a Linux System Boot-Up

To Do
debian install
hello world service


Adding the service:
navigate to the directory for the script file. Copy to directory.
ex. /usr/local/bin/project1
enable permissions
ex. chmod u+x /usr/local/bin/project1
navigate and copy the service file
ex. /lib/systemd/system/project1.service
reload daemons
ex. systemctl daemon-reload
enable service
ex. systemctl enable project1.service
you can view the systemd plot to check if the service ran
ex. systemd-analyze plot > boot.svg


Research Resources
view boot logs
https://superuser.com/questions/1081851/see-the-systemd-boot-logs

boot loader messages thru VNC console
https://www.infomaniak.com/en/support/faq/2182/displaying-the-bootloader-for-an-unmanaged-cloud-server-from-the-console

bootloader variables in systemd. Perhaps we can print these during boot somehow
https://systemd.io/BOOT_LOADER_INTERFACE/

enable viewing boot messages during boot
https://askubuntu.com/questions/248/how-can-i-show-or-hide-boot-messages-when-ubuntu-starts

remote boot messages (netconsole)
https://unix.stackexchange.com/questions/391594/is-it-possible-to-see-the-messages-that-are-displayed-at-boot-of-a-server-with-a

netconsole info
https://www.kernel.org/doc/html/latest/networking/netconsole.html
