package main

type OutputWriter interface {
	Write(data []byte, srcPort uint16, destPort uint16, srcAddr string, destAddr string) (int, error)
}
