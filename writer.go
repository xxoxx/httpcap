package main

type OutputWriter interface {
	Write(data []byte, srcPort uint16, destPort uint16, localAddr string, remoteAddr string) (int, error)
}
