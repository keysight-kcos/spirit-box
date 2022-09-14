// Networking and hardware.
package device

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

var SERVER_PORT = "8080" // spirit-box server port
var HOST_PORT = "80"     // port that host machine's default server uses
var TEMP_PORT = "8081"   // port to use redirect HOST_PORT to while waiting for that server to come up
var NIC = "eth0"         // nic to set iptables rules for
var HOST_IS_UP = false

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

func GetAddrs(interfaceName string) ([]string, error) {
	i, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	addrs, err := i.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		address := addr.String()
		if !isLinkLocalIPv6(address) {
			ret = append(ret, address)
		}
	}

	return ret, nil
}

func CreateIPStr() string {
	ips, err := GetAddrs(NIC)
	if err != nil {
		log.Print(err)
		ips = []string{"not found"}
	}
	return fmt.Sprintf("IP: %s\nPorts: host -> %s, spirit-box server -> %s", strings.Join(ips, ", "), HOST_PORT, SERVER_PORT)
}

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func isLinkLocalIPv6(address string) bool {
	return strings.HasPrefix(address, "fe80")
}

func Stub() {
	interfaceName := "mgmt0"
	i, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Fatal(err)
	}

	/*
		all, err := net.InterfaceAddrs()
		if err != nil {
			log.Fatal(err)
		}
	*/

	multi, err := i.MulticastAddrs()
	if err != nil {
		log.Fatal(err)
	}

	uni, err := i.Addrs()
	if err != nil {
		log.Fatal(err)
	}

	/*
		fmt.Printf("All:\n")
		for _, addr := range all {
			fmt.Printf("%s: %s\n", addr.String(), addr.Network())
		}
	*/

	fmt.Printf("Unicast:\n")
	for _, addr := range uni {
		address, network := addr.String(), addr.Network()
		fmt.Printf("%s: %s, IsLinkLocalIPv6: %t\n", address, network, isLinkLocalIPv6(address))
	}

	fmt.Printf("Multicast:\n")
	for _, addr := range multi {
		address, network := addr.String(), addr.Network()
		fmt.Printf("%s: %s, IsLinkLocalIPv6: %t\n", address, network, isLinkLocalIPv6(address))
	}
}

func SetRules(addFlag, nic, from, to string) error {
	// -A or -D for addFlag
	args := strings.Split(
		fmt.Sprintf("-t nat %s PREROUTING -i %s -p tcp --dport %s -j REDIRECT --to %s", addFlag, nic, from, to), " ")
	cmd := exec.Command("iptables", args...)
	bytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error setting iptables rule: %w: %s, %v", err, string(bytes), args)
	}

	args = strings.Split(
		fmt.Sprintf("-t nat %s OUTPUT -p tcp --dport %s -j REDIRECT --to %s", addFlag, from, to), " ")
	cmd = exec.Command("iptables", args...)
	bytes, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("Error setting iptables rule: %w: %s, %v", err, string(bytes), args)
	}

	return nil
}

func SetPortForwarding() error {
	var err error
	for i := 0; i < 10; i++ {
		err = SetRules("-A", NIC, TEMP_PORT, HOST_PORT)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(250) * time.Millisecond)
	}
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		err = SetRules("-A", NIC, HOST_PORT, SERVER_PORT)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(250) * time.Millisecond)
	}
	return err
}

func UnsetPortForwarding() error {
	err := SetRules("-D", NIC, HOST_PORT, SERVER_PORT)
	if err != nil {
		return err
	}
	return SetRules("-D", NIC, TEMP_PORT, HOST_PORT)
}
