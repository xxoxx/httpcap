package writer

type OutputWriter interface {
	Write(data []byte, srcPort int, destPort int, srcAddr string, destAddr string, seq uint32) (int, error)
}
