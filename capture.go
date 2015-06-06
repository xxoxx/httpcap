package main

import (
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"strings"
	"time"
)

func startCapture() {
	//_, defaultIp := GetFirstInterface()
	ipaddr := strings.Join([]string{"0.0.0.0", Setting.Port}, ":")
	if Setting.InterfaceName != "" {
		trial := net.ParseIP(Setting.InterfaceName)
		if trial.To4() == nil {
			iface, _ := net.InterfaceByName(Setting.InterfaceName)
			ipaddr = strings.Join([]string{GetIp(iface), Setting.Port}, ":")
		} else {
			ipaddr = strings.Join([]string{Setting.InterfaceName, Setting.Port}, ":")
		}
	}
	input := NewRAWInput(ipaddr)
	output := NewHttpOutput("")

	go CopyMulty(input, output)

	for {
		select {
		// case <-stop:
		// 	return
		case <-time.After(1 * time.Second):
		}
	}
}

func CopyMulty(src InputReader, writers ...OutputWriter) (err error) {
	// Don't exit on panic
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(error); !ok {
				fmt.Printf("PANIC: pkg: %v %s \n", r, debug.Stack())
			} else {
				fmt.Printf("PANIC: pkg: %s \n", debug.Stack())
			}
			CopyMulty(src, writers...)
		}
	}()

	buf := make([]byte, 5*1024*1024)
	wIndex := 0

	for {
		nr, srcPort, destPort, srcAddr, destAddr, er := src.Read(buf)
		if nr > 0 && len(buf) > nr {
			Debug("Sending", src, ": ", string(buf[0:nr]))

			if true {
				// Simple round robin
				writers[wIndex].Write(buf[0:nr], srcPort, destPort, srcAddr, destAddr)

				wIndex++

				if wIndex >= len(writers) {
					wIndex = 0
				}
			} else {
				for _, dst := range writers {
					dst.Write(buf[0:nr], srcPort, destPort, srcAddr, destAddr)
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
