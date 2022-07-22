// Networking and hardware.
package device

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"spirit-box/config"
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

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
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
		err = SetRules("-A", NIC, HOST_PORT, SERVER_PORT)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(250) * time.Millisecond)
	}
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		err = SetRules("-A", NIC, TEMP_PORT, HOST_PORT)
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

// Just loading one each for now
func LoadNetworkConfig() {
	type ParseObj struct {
		ServerPort string `json:"serverPort"`
		HostPort   string `json:"hostPort"`
		TempPort   string `json:"tempPort"`
		Nic        string `json:"nic"`
	}

	temp := ParseObj{}

	bytes, err := os.ReadFile(config.NETWORK_CONFIG_PATH)
	if err != nil { // just use defaults on error
		return
	}

	err = json.Unmarshal(bytes, &temp)
	if err != nil { // just use defaults on error
		return
	}

	SERVER_PORT = temp.ServerPort
	HOST_PORT = temp.HostPort
	TEMP_PORT = temp.TempPort
	NIC = temp.Nic
}
