package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	raw "http-sniffer/raw_socket_listener"
)

type RAWInput struct {
	data    chan RAWData
	address string
}

func NewRAWInput(address string, port string) (i *RAWInput) {
	ipaddr := strings.Join([]string{"0.0.0.0", port}, ":")
	if address != "" {
		trial := net.ParseIP(address)
		if trial.To4() == nil {
			iface, err := net.InterfaceByName(address)
			if err != nil {
				log.Fatal(err)
			}
			ipaddr = strings.Join([]string{GetIp(iface), port}, ":")
		} else {
			ipaddr = strings.Join([]string{address, port}, ":")
			address = ipaddr
		}
	} else {
		address = "0.0.0.0"
	}

	if port == "" {
		fmt.Printf("listen on %s\n\n", address)
	} else {
		fmt.Printf("listen on %s:%s\n\n", address, port)
	}

	i = new(RAWInput)
	i.data = make(chan RAWData)
	i.address = ipaddr

	go i.listen(ipaddr)

	return
}

func (i *RAWInput) Read(data []byte) (int, uint16, uint16, string, string, error) {
	raw := <-i.data
	copy(data, raw.Data)

	return len(raw.Data), raw.SrcPort, raw.DestPort, raw.SrcAddr, raw.DestAddr, nil
}

func (i *RAWInput) listen(address string) {
	address = strings.Replace(address, "[::]", "127.0.0.1", -1)

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		log.Fatal("input-raw: error while parsing address", err)
	}

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

func (i *RAWInput) String() string {
	return "RAW Socket input: " + i.address
}
