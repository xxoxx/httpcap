package raw_socket

import (
	_ "fmt"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
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

	la := &syscall.SockaddrInet4{}
	if err := syscall.Bind(fd, la); err != nil {
		log.Fatalln("Error:Bind", err)
	}

	var addr *net.IPAddr
	var src_ip string
	var dest_ip string
	var flags uint32
	var qty uint32

	buf := make([]byte, 65536)
	wbuf := syscall.WSABuf{Buf: &buf[0], Len: uint32(len(buf))}
	rsa := new(syscall.RawSockaddrAny)
	rsan := int32(unsafe.Sizeof(*rsa))
	// o := syscall.Overlapped{}

	for {

		// http://www.binarytides.com/packet-sniffer-code-in-c-using-winsock/
		err = syscall.WSARecvFrom(fd, &wbuf, 1, &qty, &flags, rsa, &rsan, nil, nil)
		if err != nil {
			log.Println("Error:WSARecvFrom", err)
			continue
		}
		if qty <= 0 {
			continue
		}

		packet := gopacket.NewPacket(buf[:qty], layers.LayerTypeIPv4, gopacket.Default)
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)

			addr = &net.IPAddr{}
			src_ip = packet.NetworkLayer().NetworkFlow().Src().String()
			dest_ip = packet.NetworkLayer().NetworkFlow().Dst().String()
			n := len(tcp.Contents) + len(tcp.Payload)
			if len(tcp.Contents) > 0 {
				copy(buf, tcp.Contents)
			}
			if len(tcp.Payload) > 0 {
				copy(buf[len(tcp.Contents):], tcp.Payload)
			}

			t.parsePacket(addr, src_ip, dest_ip, buf[:n])
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
