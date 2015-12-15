package raw_socket

import (
	_ "fmt"
	_ "os"
	"strconv"

	"github.com/cxfksword/httpcap/common"
	"github.com/google/gopacket/layers"
)

const (
	IP_HDRINCL = 2
)

// Capture traffic from socket using RAW_SOCKET's
// http://en.wikipedia.org/wiki/Raw_socket
//
// RAW_SOCKET allow you listen for traffic on any port (e.g. sniffing) because they operate on IP level.
// Ports is TCP feature, same as flow control, reliable transmission and etc.
// Since we can't use default TCP libraries RAWTCPLitener implements own TCP layer
// TCP packets is parsed using tcp_packet.go, and flow control is managed by tcp_message.go
type Listener struct {
	messages map[string]*TCPMessage // buffer of TCPMessages waiting to be send

	c_packets  chan *TCPPacket
	c_messages chan *TCPMessage // Messages ready to be send to client

	c_del_message chan *TCPMessage // Used for notifications about completed or expired messages

	addr string // IP to listen
	port int    // Port to listen
	host string
}

// RAWTCPListen creates a listener to capture traffic from RAW_SOCKET
func NewListener(addr string, port string) (rawListener *Listener) {
	rawListener = &Listener{}

	rawListener.c_packets = make(chan *TCPPacket, 100)
	rawListener.c_messages = make(chan *TCPMessage, 100)
	rawListener.c_del_message = make(chan *TCPMessage, 100)
	rawListener.messages = make(map[string]*TCPMessage)

	rawListener.addr = addr
	rawListener.port, _ = strconv.Atoi(port)
	rawListener.host = common.GetHostIp()

	go rawListener.listen()
	go rawListener.readRAWSocket()

	return
}

func (t *Listener) listen() {
	for {
		select {
		// If message ready for deletion it means that its also complete or expired by timeout
		case message := <-t.c_del_message:
			t.c_messages <- message
			delete(t.messages, message.ID)

		// We need to use channels to process each packet to avoid data races
		case packet := <-t.c_packets:
			t.processTCPPacket(packet)
		}
	}
}

func (t *Listener) parsePacket(srcIp string, destIp string, packet *layers.TCP) {
	if t.isListenPacket(srcIp, destIp, packet) {
		t.c_packets <- ParseTCPPacket(srcIp, destIp, packet)
	}
}

func (t *Listener) isListenPacket(srcIp string, destIp string, packet *layers.TCP) bool {
	// filter SYN,FIN,ACK-only packets not have data inside and Keepalive hearbeat packets with no data inside
	if len(packet.Payload) == 0 {
		return false
	}

	// filter Keepalive hearbeat packets with 1-byte segment on Windows
	if packet.ACK && len(packet.Payload) == 1 {
		return false
	}

	// listen all port packet
	if t.port <= 0 {
		return true
	}

	// Because RAW_SOCKET can't be bound to port, we have to control it by ourself
	if (t.host == srcIp || srcIp == "127.0.0.1") && int(packet.SrcPort) == t.port {
		return true
	}

	if (t.host == destIp || destIp == "127.0.0.1") && int(packet.DstPort) == t.port {
		return true
	}

	return false
}

// Trying to add packet to existing message or creating new message
//
// For TCP message unique id is Acknowledgment number (see tcp_packet.go)
func (t *Listener) processTCPPacket(packet *TCPPacket) {
	defer func() { recover() }()

	var message *TCPMessage
	m_id := packet.SrcIP + strconv.Itoa(int(packet.tcp.Ack))

	message, ok := t.messages[m_id]

	if !ok {
		// We sending c_del_message channel, so message object can communicate with Listener and notify it if message completed
		message = NewTCPMessage(m_id, t.c_del_message)
		t.messages[m_id] = message
	}

	// Adding packet to message
	message.c_packets <- packet
}

// Receive TCP messages from the listener channel
func (t *Listener) Receive() *TCPMessage {
	return <-t.c_messages
}
