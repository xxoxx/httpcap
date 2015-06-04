package main

import (
	_ "fmt"
	"io"
	"net"
	"strings"
	"time"
)

func startCapture() {
	_, defaultIp := GetFirstInterface()
	ipaddr := strings.Join([]string{defaultIp, Setting.Port}, ":")
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

func CopyMulty(src io.Reader, writers ...io.Writer) (err error) {
	buf := make([]byte, 5*1024*1024)
	wIndex := 0

	for {
		nr, er := src.Read(buf)
		if nr > 0 && len(buf) > nr {
			Debug("Sending", src, ": ", string(buf[0:nr]))

			if true {
				// Simple round robin
				writers[wIndex].Write(buf[0:nr])

				wIndex++

				if wIndex >= len(writers) {
					wIndex = 0
				}
			} else {
				for _, dst := range writers {
					dst.Write(buf[0:nr])
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
