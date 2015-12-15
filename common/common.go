package common

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/cxfksword/httpcap/config"
)

type ByLength []string

func (s ByLength) Len() int {
	return len(s)
}
func (s ByLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByLength) Less(i, j int) bool {
	return len(s[i]) > len(s[j])
}

func ShowAllInterfaces() {
	ifaces, _ := net.Interfaces()

	iplist := ""
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()

		ipV4 := false
		ipAddrs := []string{}
		for _, addr := range addrs {
			var ip net.IP
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip = ipnet.IP
			} else if ipaddr, ok := addr.(*net.IPAddr); ok {
				ip = ipaddr.IP
			}
			if ip != nil && ip.To4() != nil && !ip.IsUnspecified() {
				ipstr := addr.String()
				idx := strings.Index(ipstr, "/")
				if idx >= 0 {
					ipstr = ipstr[:idx]
				}
				ipAddrs = append(ipAddrs, ipstr)
				ipV4 = true
			}
		}
		if !ipV4 {
			continue
		}

		iplist += fmt.Sprintf("%-7d %-40s %s\n", iface.Index, iface.Name, strings.Join(ipAddrs, ", "))
	}

	fmt.Printf("%-7s %-40s %s\n", "index", "interface name", "ip")
	fmt.Print(iplist)
}

func GetHostIp() string {
	ip := "127.0.0.1"

	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	}

	return ip
}

func GetFirstInterface() (name string, ip string) {
	ifaces, _ := net.Interfaces()

	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()

		ipV4 := false
		for _, addr := range addrs {
			var ip net.IP
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip = ipnet.IP
			} else if ipaddr, ok := addr.(*net.IPAddr); ok {
				ip = ipaddr.IP
			}
			if ip != nil && ip.To4() != nil && !ip.IsUnspecified() {
				ipstr := addr.String()
				idx := strings.Index(ipstr, "/")
				if idx >= 0 {
					ipstr = ipstr[:idx]
				}

				return iface.Name, ipstr
			}
		}
		if !ipV4 {
			continue
		}

	}

	return "", "0.0.0.0"
}

func Debug(args ...interface{}) {
	if config.Setting.Verbose {
		log.Println(args...)
	}
}
