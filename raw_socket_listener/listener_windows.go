package raw_socket

import (
	_ "fmt"
	"log"
	"net"
)

func (t *Listener) readRAWSocket() {

	conn, e := net.ListenPacket("ip4", t.addr)
	if e != nil {
		log.Fatal(e)
	}
	defer conn.Close()

	var n int
	var addr *net.IPAddr
	var err error
	var src_ip string
	var dest_ip string

	buf := make([]byte, 4096*2)
	hostIp := getHostIP()

	for {

		// Note: windows raw socket for the IPPROTO_TCP protocol is not allowed
		// https://msdn.microsoft.com/en-us/library/windows/desktop/ms740548%28v=vs.85%29.aspx
		// ReadFromIP receive messages without IP header
		n, addr, err = conn.(*net.IPConn).ReadFromIP(buf)
		// TODO: judge windows incoming/outgoing package not accurate, maybe replace with winpcap.
		if addr.String() == hostIp {
			// outgoing package
			src_ip = addr.String()
			dest_ip = "0.0.0.0" // can't get dest ip
		} else {
			// incoming package
			src_ip = addr.String()
			dest_ip = hostIp
		}

		if err != nil {
			log.Println("Error:", err)
			continue
		}

		if n > 0 {
			t.parsePacket(addr, src_ip, dest_ip, buf[:n])
		}
	}

}
