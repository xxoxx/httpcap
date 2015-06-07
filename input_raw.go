package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"

	raw "github.com/cxfksword/httpcap/raw_socket_listener"
)

type RAWInput struct {
	data    chan RAWData
	address string
}

func NewRAWInput(host string, port string) (i *RAWInput) {
	ifname := ""
	if host != "" && host != "0.0.0.0" {
		trial := net.ParseIP(host)
		if trial.To4() == nil {
			// host is interface name
			ifname = host
			iface, err := net.InterfaceByName(ifname)
			if err != nil {
				log.Fatal(err)
			}
			host = GetIp(iface)
		} else {
			// host is ip address
			ifname = GetInterfaceNameByIp(trial)
		}
	} else {
		host = "0.0.0.0"
		ifname = "0.0.0.0"
	}

	i = new(RAWInput)
	i.data = make(chan RAWData)
	i.address = host

	if runtime.GOOS == "windows" {
		go i.listen(host, port)
		if port == "" {
			fmt.Printf("listen on %s\n\n", host)
		} else {
			fmt.Printf("listen on %s:%s\n\n", host, port)
		}
	} else {
		go i.listen(ifname, port)
		if port == "" {
			fmt.Printf("listen on %s\n\n", ifname)
		} else {
			fmt.Printf("listen on %s:%s\n\n", ifname, port)
		}
	}

	return
}

func (i *RAWInput) Read(data []byte) (int, uint16, uint16, string, string, error) {
	raw := <-i.data
	copy(data, raw.Data)

	return len(raw.Data), raw.SrcPort, raw.DestPort, raw.SrcAddr, raw.DestAddr, nil
}

func (i *RAWInput) listen(host string, port string) {
	host = strings.Replace(host, "[::]", "127.0.0.1", -1)

	listener := raw.NewListener(host, port)

	for {
		// Receiving TCPMessage object
		m := listener.Receive()

		i.data <- RAWData{
			Data:      m.Bytes(),
			SrcPort:   m.SourcePort(),
			DestPort:  m.DestinationPort(),
			LocalAddr: i.address,
			SrcAddr:   m.SourceIP(),
			DestAddr:  m.DestinationIP(),
		}
	}
}

func GetFirstInterface() (name string, ip string) {
	ifaces, _ := net.Interfaces()
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
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() {
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

		return iface.Name, ipAddrs[0]
	}

	return "", "0.0.0.0"
}

func GetIp(iface *net.Interface) string {
	addrs, _ := iface.Addrs()

	ipAddrs := []string{}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPAddr); ok && !ip.IP.IsUnspecified() {
			ipAddrs = append(ipAddrs, addr.String())
		}
	}

	if len(ipAddrs) > 0 {
		return ipAddrs[0]
	} else {
		return ""
	}
}

func GetInterfaceNameByIp(checkAddr net.IP) string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()

		for _, addr := range addrs {
			var ip net.IP
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip = ipnet.IP
			} else if ipaddr, ok := addr.(*net.IPAddr); ok {
				ip = ipaddr.IP
			}
			if ip != nil && ip.To4().String() == checkAddr.String() {
				return iface.Name
			}
		}
	}

	return "0.0.0.0"
}

func (i *RAWInput) String() string {
	return "RAW Socket input: " + i.address
}
