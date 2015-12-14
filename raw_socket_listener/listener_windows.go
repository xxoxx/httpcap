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

			// pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
			// sa := new(syscall.SockaddrInet4)
			// p := (*[2]byte)(unsafe.Pointer(&pp.Port))
			// sa.Port = int(p[0])<<8 + int(p[1])
			// for i := 0; i < len(sa.Addr); i++ {
			// 	sa.Addr[i] = pp.Addr[i]
			// }
			// remoteAddr := &net.IPAddr{IP: net.IPv4(sa.Addr[0], sa.Addr[1], sa.Addr[2], sa.Addr[3])}

			src_ip = packet.NetworkLayer().NetworkFlow().Src().String()
			dest_ip = packet.NetworkLayer().NetworkFlow().Dst().String()
			remoteAddr := &net.IPAddr{IP: net.ParseIP(src_ip)}
			n := len(tcp.Contents) + len(tcp.Payload)

			// log.Println(src_ip + " -> " + dest_ip)

			if len(tcp.Contents) > 0 {
				copy(buf, tcp.Contents)
			}
			if len(tcp.Payload) > 0 {
				copy(buf[len(tcp.Contents):], tcp.Payload)
			}

			t.parsePacket(remoteAddr, src_ip, dest_ip, buf[:n])
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
