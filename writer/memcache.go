package writer

import (
  "fmt"
  "strings"
)

type MemcacheOutput struct {
}

func NewMemcacheOutput(options string) (di *MemcacheOutput) {
	di = new(MemcacheOutput)

	return
}

func (i *MemcacheOutput) Write(data []byte, srcPort int, destPort int, srcAddr string, destAddr string, isOutputPacket bool) (int, error) {
    	cmd := string(data[:3])
        switch cmd {
		case "get":
			fallthrough
		case "set":
			idx := strings.Index(string(data), "\n")
                        cmdstr := string(data[:idx])
                        fmt.Println("[MC]" + srcAddr + " -> " + destAddr + " " + cmdstr)		
        }

	return 0, nil
}
