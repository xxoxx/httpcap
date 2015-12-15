package writer

type MemcacheOutput struct {
}

func NewMemcacheOutput(options string) (di *MemcacheOutput) {
	di = new(MemcacheOutput)

	return
}

func (i *MemcacheOutput) Write(data []byte, srcPort int, destPort int, srcAddr string, destAddr string, isOutputPacket bool) (int, error) {
	return 0, nil
}
