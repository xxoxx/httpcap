package raw_socket

import (
	"github.com/google/gopacket/layers"
)

type TCPPacket struct {
	tcp *layers.TCP

	SrcIP  string
	DestIP string
}

func ParseTCPPacket(src_ip string, dest_ip string, packet *layers.TCP) (p *TCPPacket) {
	p = &TCPPacket{}
	p.SrcIP = src_ip
	p.DestIP = dest_ip
	p.tcp = packet

	return p
}

// String output for a TCP Packet
// func (t *TCPPacket) String() string {
// 	return strings.Join([]string{
// 		"Source port: " + strconv.Itoa(int(t.SrcPort)),
// 		"Dest port:" + strconv.Itoa(int(t.DestPort)),
// 		"Sequence:" + strconv.Itoa(int(t.Seq)),
// 		"Acknowledgment:" + strconv.Itoa(int(t.Ack)),
// 		"Header len:" + strconv.Itoa(int(t.DataOffset)),

// 		"Flag ns:" + strconv.FormatBool(t.Flags&TCP_NS != 0),
// 		"Flag crw:" + strconv.FormatBool(t.Flags&TCP_CWR != 0),
// 		"Flag ece:" + strconv.FormatBool(t.Flags&TCP_ECE != 0),
// 		"Flag urg:" + strconv.FormatBool(t.Flags&TCP_URG != 0),
// 		"Flag ack:" + strconv.FormatBool(t.Flags&TCP_ACK != 0),
// 		"Flag psh:" + strconv.FormatBool(t.Flags&TCP_PSH != 0),
// 		"Flag rst:" + strconv.FormatBool(t.Flags&TCP_RST != 0),
// 		"Flag syn:" + strconv.FormatBool(t.Flags&TCP_SYN != 0),
// 		"Flag fin:" + strconv.FormatBool(t.Flags&TCP_FIN != 0),

// 		"Window size:" + strconv.Itoa(int(t.Window)),
// 		"Checksum:" + strconv.Itoa(int(t.Checksum)),

// 		// "Data:" + string(t.Data),
// 	}, "\n")
// }

type BySeq []*TCPPacket

func (a BySeq) Len() int           { return len(a) }
func (a BySeq) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeq) Less(i, j int) bool { return a[i].tcp.Seq < a[j].tcp.Seq }
