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

func (i *HttpOutput) Write(data []byte) (int, error) {
	fmt.Println("##################")
	buf := bytes.NewBuffer(data)
	reader := bufio.NewReader(buf)
	request, err = http.ReadRequest(reader)

	if err == nil {
		if request.Method == "POST" {
			body, _ := ioutil.ReadAll(reader)
			bodyBuf := bytes.NewBuffer(body)
			request.Body = ioutil.NopCloser(bodyBuf)
			request.ContentLength = int64(bodyBuf.Len())
		}

		fmt.Println(i.ReadRawHeader(data))
	} else {
		buf := bytes.NewBuffer(data)
		reader := bufio.NewReader(buf)
		response, err := http.ReadResponse(reader, request)
		if err != nil {
			log.Printf("Can't parse request data. %s\n", err.Error())
			return 0, nil
		}

		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			r, _ := gzip.NewReader(response.Body)
			defer r.Close()
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
		fmt.Println(i.ReadRawHeader(data))
	}
	//fmt.Println("Writing message: #", string(data))

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

func (i *HttpOutput) String() string {
	return "Dummy Output"
}
