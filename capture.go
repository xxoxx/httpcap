package main

import (
	"fmt"
	"io"
	"runtime/debug"
	"time"
        "log"

	"github.com/cxfksword/httpcap/common"
	"github.com/cxfksword/httpcap/config"
	"github.com/cxfksword/httpcap/reader"
	"github.com/cxfksword/httpcap/writer"
)

func startCapture() {
	input := reader.NewRAWInput(config.Setting.InterfaceName, config.Setting.Port)

	go CopyMulty(input)

	for {
		select {
		// case <-stop:
		// 	return
		case <-time.After(1 * time.Second):
		}
	}
}

func CopyMulty(src reader.InputReader) (err error) {
	// Don't exit on panic
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(error); !ok {
				fmt.Printf("PANIC: pkg: %v %s \n", r, debug.Stack())
			} else {
				fmt.Printf("PANIC: pkg: %s \n", debug.Stack())
			}
                        log.Fatal(r.(error))
		}
	}()

	http := writer.NewHttpOutput("")
	memcache := writer.NewMemcacheOutput("")

	services := common.DiscoverServices()
	buf := make([]byte, 5*1024*1024)

	for {
		nr, srcPort, destPort, srcAddr, destAddr, er := src.Read(buf)
		if nr > 0 && len(buf) > nr {
			common.Debug("Sending", src, ": ", string(buf[0:nr]))

			servicePort := destPort
			isOutputPacket := false
			if ip := common.GetHostIp(); ip == srcAddr {
				isOutputPacket = true
				servicePort = srcPort
			}
                        
                        
			if srv, found := services[int(servicePort)]; found {
				switch srv.Type {
				case common.Service_Type_Memcache:
					if config.Setting.Service == "" || config.Setting.Service == "memcache" {
					memcache.Write(buf[0:nr], int(srcPort), int(destPort), srcAddr, destAddr, isOutputPacket)
                                        }

				}
			} else {
                                if !(config.Setting.Service != "" && config.Setting.Service != "http") {
				http.Write(buf[0:nr], int(srcPort), int(destPort), srcAddr, destAddr, isOutputPacket)
                               }
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return err
}
