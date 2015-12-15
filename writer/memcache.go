package writer

import (
	"fmt"
	"strings"

	"github.com/cxfksword/httpcap/color"
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
	cmd := string(data[:4])
	switch cmd {
	case "get ":
		fallthrough
	case "set ":
		fallthrough
	case "incr":
		fallthrough
	case "decr":
		idx := strings.Index(string(data), "\n")
		cmdstr := string(data[:idx])
		fmt.Println("[MC]" + srcAddr + " -> " + destAddr + " " + cmdstr)
	default:
		//if req, found := i.requests[seq]; found {
		//	size := strconv.Itoa(len(data))
		//	fmt.Println("[MC]" + req.srcAddr + " -> " + req.destAddr + " " + req.cmd + " size:" + size + "B")
		//}

		if len(data) > 5 {
			resCmd := string(data[:5])
			switch resCmd {
			case "VALUE":
				idx := strings.Index(string(data), "\n")
				cmdstr := string(data[6 : idx-1])
				arr := strings.Split(cmdstr, " ")
				color.Println("[MC]"+srcAddr+" -> "+destAddr+" get "+arr[0]+" size: "+arr[2]+"B", color.Cyan)
			default:
			}
		}
	}

	return 0, nil
}
