package raw_socket

import (
	_ "fmt"
	"log"
	"net"
	_ "strconv"
	_ "syscall"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
)

func (t *Listener) readRAWSocket() {

	// proto := (syscall.ETH_P_ALL<<8)&0xff00 | syscall.ETH_P_ALL>>8
	// AF_INET can't capture outgoing packets, must change to use AF_PACKET
	// https://github.com/golang/go/issues/7653
	// http://www.binarytides.com/packet-sniffer-code-in-c-using-linux-sockets-bsd-part-2/
	tp, e := afpacket.NewTPacket(afpacket.SocketRaw)
	if e != nil {
		log.Fatalln("Error:", e)
		return
	}
	defer tp.Close()

	var addr *net.IPAddr
	var src_ip string
	var dest_ip string
	tcpBuf := make([]byte, 4096*2)

	for {
		buf, _, err := tp.ReadPacketData()
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		if len(buf) <= 0 {
			continue
		}

		packet := gopacket.NewPacket(buf, layers.LayerTypeEthernet, gopacket.Default)
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)

			addr = &net.IPAddr{}
			src_ip = packet.NetworkLayer().NetworkFlow().Src().String()
			dest_ip = packet.NetworkLayer().NetworkFlow().Dst().String()
			n := len(tcp.Contents) + len(tcp.Payload)
			if len(tcp.Contents) > 0 {
				copy(tcpBuf, tcp.Contents)
			}
			if len(tcp.Payload) > 0 {
				copy(tcpBuf[len(tcp.Contents):], tcp.Payload)
			}

			t.parsePacket(addr, src_ip, dest_ip, tcpBuf[:n])
		}

	}

}
