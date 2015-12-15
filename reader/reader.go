package reader

type RAWData struct {
	Data      []byte
	SrcPort   uint16
	DestPort  uint16
	LocalAddr string
	SrcAddr   string
	DestAddr  string
	Seq       uint32
}

type InputReader interface {
	Read(data []byte) (int, RAWData, error)
}
