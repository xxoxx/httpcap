package raw_socket

import (
	_ "fmt"
	"log"
	"net"
	_ "strconv"
	"syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func (t *Listener) readRAWSocket() {
	// AF_INET can't capture outgoing packets, must change to use AF_PACKET
	// https://github.com/golang/go/issues/7653
	// http://www.binarytides.com/packet-sniffer-code-in-c-using-linux-sockets-bsd-part-2/
	proto := (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8 // change to Big-Endian order
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, proto)
	if err != nil {
		log.Fatal("socket: ", err)
	}
	defer syscall.Close(fd)
	if t.addr != "" && t.addr != "0.0.0.0" {
		ifi, err := net.InterfaceByName(t.addr)
		if err != nil {
			log.Fatal("interfacebyname: ", err)
		}
		lla := syscall.SockaddrLinklayer{Protocol: uint16(proto), Ifindex: ifi.Index}
		if err := syscall.Bind(fd, &lla); err != nil {
			log.Fatal("bind: ", err)
		}
	}

	var src_ip string
	var dest_ip string
	buf := make([]byte, 65536)

	for {
		n, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		if n <= 0 {
			continue
		}

		packet := gopacket.NewPacket(buf[:n], layers.LayerTypeEthernet, gopacket.Default)
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)

			src_ip = packet.NetworkLayer().NetworkFlow().Src().String()
			dest_ip = packet.NetworkLayer().NetworkFlow().Dst().String()

			t.parsePacket(src_ip, dest_ip, tcp)
		}

	}

}
