package raw_socket

import (
	"log"
	"net"
	"os"
	"syscall"
	"time"

	"github.com/cxfksword/httpcap/common"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const SIO_RCVALL = syscall.IOC_IN | syscall.IOC_VENDOR | 1

func (t *Listener) readRAWSocket() {

	var d syscall.WSAData
	err := syscall.WSAStartup(uint32(0x202), &d)
	if err != nil {
		log.Fatalln("Error:WSAStartup", err)
	}
	fd, err := sysSocket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_IP)
	if err != nil {
		log.Fatalln("Error:socket", err)
	}
	defer syscall.Close(fd)

	ip := net.ParseIP(t.addr)
	if len(t.addr) == 0 || t.addr == "0.0.0.0" {
		ip = net.IPv4zero
	}
	if ip = ip.To4(); ip == nil {
		log.Fatalln("Error: non-IPv4 address " + t.addr)
	}
	la := new(syscall.SockaddrInet4)
	for i := 0; i < net.IPv4len; i++ {
		la.Addr[i] = ip[i]
	}
	if err := syscall.Bind(fd, la); err != nil {
		log.Fatalln("Error:Bind", err)
	}

	var snapshotLen int32 = 1024
	var promiscuous bool = false
	var timeout time.Duration = 30 * time.Second

	inet := getInterfaceName()
	log.Println(inet)
	handle, err := pcap.OpenLive(inet, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)

			src_ip := packet.NetworkLayer().NetworkFlow().Src().String()
			dest_ip := packet.NetworkLayer().NetworkFlow().Dst().String()

			t.parsePacket(src_ip, dest_ip, tcp)
		}
	}
}

func sysSocket(family, sotype, proto int) (syscall.Handle, error) {
	// See ../syscall/exec_unix.go for description of ForkLock.
	syscall.ForkLock.RLock()
	s, err := syscall.Socket(family, sotype, proto)
	if err == nil {
		syscall.CloseOnExec(s)
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		return syscall.InvalidHandle, os.NewSyscallError("socket", err)
	}
	return s, nil
}

func getInterfaceName() string {
	ip := common.GetHostIp()
	log.Println(ip)
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, dev := range devices {
		addrs := dev.Addresses
		for _, addr := range addrs {
			log.Println(addr.IP.String())
			if addr.IP.String() == ip {
				return dev.Name
			}
		}
	}

	return ""
}
