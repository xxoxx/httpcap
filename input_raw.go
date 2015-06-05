package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	raw "http-sniffer/raw_socket_listener"
)

type RAWInput struct {
	data    chan RAWData
	address string
}

func NewRAWInput(address string) (i *RAWInput) {
	i = new(RAWInput)
	i.data = make(chan RAWData)
	i.address = address

	go i.listen(address)

	return
}

func (i *RAWInput) Read(data []byte) (int, uint16, uint16, string, string, error) {
	raw := <-i.data
	copy(data, raw.Data)

	return len(raw.Data), raw.SrcPort, raw.DestPort, raw.LocalAddr, raw.RemoteAddr, nil
}

func (i *RAWInput) listen(address string) {
	address = strings.Replace(address, "[::]", "127.0.0.1", -1)

	host, port, err := net.SplitHostPort(address)

	listen_port, _ := strconv.Atoi(port)
	if listen_port <= 0 {
		fmt.Printf("listen on %s\n", host)
	} else {
		fmt.Printf("listen on %s\n", address)
	}

	if err != nil {
		log.Fatal("input-raw: error while parsing address", err)
	}

	listener := raw.NewListener(host, port)

	for {
		// Receiving TCPMessage object
		m := listener.Receive()

		i.data <- RAWData{
			Data:       m.Bytes(),
			SrcPort:    m.SourcePort(),
			DestPort:   m.DestinationPort(),
			LocalAddr:  i.address,
			RemoteAddr: m.RemoteAddress(),
		}
	}
}

func (i *RAWInput) String() string {
	return "RAW Socket input: " + i.address
}
