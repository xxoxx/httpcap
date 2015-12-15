package writer

import (
	"fmt"
	"strconv"
	"strings"
)

type MemcacheOutput struct {
	requests map[uint32]*memcacheData
}

type memcacheData struct {
	cmd      string
	srcPort  int
	destPort int
	srcAddr  string
	destAddr string
}

func NewMemcacheOutput(options string) (di *MemcacheOutput) {
	di = new(MemcacheOutput)
	di.requests = make(map[uint32]*memcacheData)
	return
}

func (i *MemcacheOutput) Write(data []byte, srcPort int, destPort int, srcAddr string, destAddr string, seq uint32) (int, error) {
	cmd := string(data[:3])
	switch cmd {
	case "get":
		fallthrough
	case "set":
		idx := strings.Index(string(data), "\n")
		cmdstr := string(data[:idx])
		fmt.Println("[MC]" + srcAddr + " -> " + destAddr + " " + cmdstr)
		i.requests[seq] = &memcacheData{cmd: cmdstr, srcPort: srcPort, destPort: destPort, srcAddr: srcAddr, destAddr: destAddr}
	default:
		if req, found := i.requests[seq]; found {
			size := strconv.Itoa(len(data))
			fmt.Println("[MC]" + req.srcAddr + " -> " + req.destAddr + " " + req.cmd + " size:" + size + "B")
		}
	}

	return 0, nil
}
