// Networking and hardware.
package device

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func PrintInterfaces() {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range interfaces {
		var ips []string
		ipv4_exists := false // keep checking until interface has ipv4 address
		for {
			ips = make([]string, 0)
			addrs, err := i.Addrs()
			if err != nil {
				log.Fatal(err)
			}
			for _, addr := range addrs {
				str := addr.String()
				ips = append(ips, str)
				ipv4_exists = ipv4_exists || isIPv4(str)
			}
			if ipv4_exists {
				break
			}
		}
		out := fmt.Sprintf("%s: %s\n", i.Name, strings.Join(ips, ", "))
		fmt.Print(out)
	}
	fmt.Println()
}

func GetIPv4Addr(interfaceName string) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range interfaces {
		if i.Name != interfaceName {
			continue
		}

		for { // IPs are not always available immediately. unsafe workaround that should be changed.
			addrs, err := i.Addrs()
			if err != nil {
				log.Fatal(err)
			}
			for _, addr := range addrs {
				str := addr.String()
				if isIPv4(str) {
					return str
				}
			}
		}
	}
	return "Unable to find IP."
}

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}
