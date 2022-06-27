// Networking and hardware.
package device

import (
	"spirit-box/logging"
	"log"
	"fmt"
	"net"
	"strings"
)

func PrintInterfaces() {
	l := logging.Logger

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal(err)
		}

		ips := make([]string, 0)
		for _, addr := range addrs {
			str := addr.String()
			if isIPv4(str) {
				ips = append(ips, str)
			}
		}
		out := fmt.Sprintf("%s: %s\n", i.Name, strings.Join(ips, ","))
		fmt.Print(out)
		l.Print(out)
	}
	fmt.Println()
}

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}
