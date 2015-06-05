package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"strings"
)

type HttpOutput struct {
}

func NewHttpOutput(options string) (di *HttpOutput) {
	di = new(HttpOutput)

	return
}

var request *http.Request
var err error
var count uint

func (i *HttpOutput) Write(data []byte, srcPort uint16, destPort uint16, localAddr string, remoteAddr string) (int, error) {
	if i.isHttps(srcPort, destPort) {
		fmt.Println("ssl request")
		return 0, nil
	}
	if i.isRequest(data) {
		buf := bytes.NewBuffer(data)
		reader := bufio.NewReader(buf)
		request, err = http.ReadRequest(reader)
		if err != nil {
			log.Printf("Can't parse request data. %s\n", err.Error())
			return 0, nil
		}

		if request.Method == "POST" || request.Method == "PUT" {
			body, _ := ioutil.ReadAll(reader)
			bodyBuf := bytes.NewBuffer(body)
			request.Body = ioutil.NopCloser(bodyBuf)
			request.ContentLength = int64(bodyBuf.Len())
		}

		count++
		// fmt.Println(i.ReadRawHeader(data))

		fmt.Printf("%-5d %-7s %-5s %s\n", count, request.Method, "-", "http://"+request.Host+request.RequestURI)
	}

	if i.isResponse(data) {
		buf := bytes.NewBuffer(data)
		reader := bufio.NewReader(buf)
		response, err := http.ReadResponse(reader, request)
		if err != nil {
			log.Printf("Can't parse response data. %s\n", err.Error())
			return 0, nil
		}

		if !i.allowShowResponseBody(response.Header.Get("Content-Type")) {
			return 0, nil
		}

		fmt.Println(i.ReadRawHeader(data))
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			r, err := gzip.NewReader(response.Body)
			if err == io.EOF {
				return 0, nil
			}
			body, _ := ioutil.ReadAll(r)
			bodyBuf := bytes.NewBuffer(body)
			response.Body = ioutil.NopCloser(bodyBuf)
			response.ContentLength = int64(bodyBuf.Len())
		default:
			body, _ := ioutil.ReadAll(reader)
			bodyBuf := bytes.NewBuffer(body)
			response.Body = ioutil.NopCloser(bodyBuf)
			response.ContentLength = int64(bodyBuf.Len())
		}

		fmt.Println(response.Body)
	}

	return len(data), nil
}

func (i *HttpOutput) ReadRawHeader(data []byte) string {
	buf := bytes.NewBuffer(data)
	reader := bufio.NewReader(buf)
	tp := textproto.NewReader(reader)

	headers := ""
	for line, err := tp.ReadLine(); err != io.EOF; {
		if strings.TrimSpace(line) == "" {
			break
		} else {
			headers += line + "\n"
		}

		line, err = tp.ReadLine()
	}

	return strings.TrimSpace(headers)
}

func (i *HttpOutput) isRequest(data []byte) bool {
	buf := bytes.NewBuffer(data)
	reader := bufio.NewReader(buf)
	tp := textproto.NewReader(reader)

	firstLine, _ := tp.ReadLine()
	arr := strings.Split(firstLine, " ")

	switch strings.TrimSpace(arr[0]) {
	case "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT":
		return true
	default:
		return false
	}
}

func (i *HttpOutput) isResponse(data []byte) bool {
	buf := bytes.NewBuffer(data)
	reader := bufio.NewReader(buf)
	tp := textproto.NewReader(reader)

	firstLine, _ := tp.ReadLine()
	return strings.HasPrefix(strings.TrimSpace(firstLine), "HTTP/")
}

func (i *HttpOutput) isHttps(srcPort uint16, destPort uint16) bool {
	return srcPort == 443 || destPort == 443
}

func (i *HttpOutput) allowShowResponseBody(contentType string) bool {
	contentType = strings.ToLower(contentType)
	switch {
	case strings.Contains(contentType, "text/html"),
		strings.Contains(contentType, "application/json"),
		strings.Contains(contentType, "text/javascript"),
		strings.Contains(contentType, "text/css"):
		return true
	}

	return false
}

func (i *HttpOutput) String() string {
	return "Dummy Output"
}
