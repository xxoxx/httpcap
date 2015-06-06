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
	"runtime"
	"strings"
	"sync"
	"time"
)

type HttpOutput struct {
	requests map[string]*httpRequestData
}

type httpRequestData struct {
	request  *http.Request
	srcPort  uint16
	destPort uint16
	srcAddr  string
	destAddr string
	header   string
	addTime  time.Time
}

func NewHttpOutput(options string) (di *HttpOutput) {
	di = new(HttpOutput)
	di.requests = make(map[string]*httpRequestData)

	return
}

var request *http.Request
var err error
var count uint
var locker *sync.Mutex = &sync.Mutex{}

func (i *HttpOutput) Write(data []byte, srcPort uint16, destPort uint16, srcAddr string, destAddr string) (int, error) {
	if i.isHttps(srcPort, destPort) {
		// TODO: can't get CONNECT request
		// https package
		return 0, nil
	}
	if i.isRequest(data) {
		buf := bytes.NewBuffer(data)
		reader := bufio.NewReader(buf)

		// read header
		request, err = http.ReadRequest(reader)
		if err != nil {
			log.Printf("Can't parse request data. %s\n", err.Error())
			return 0, nil
		}

		// read body
		if request.Method == "POST" || request.Method == "PUT" {
			body, _ := ioutil.ReadAll(reader)
			bodyBuf := bytes.NewBuffer(body)
			request.Body = ioutil.NopCloser(bodyBuf)
			request.ContentLength = int64(bodyBuf.Len())
		}

		requestData := httpRequestData{
			request:  request,
			header:   i.ReadRawHeader(data),
			srcPort:  srcPort,
			destPort: destPort,
			srcAddr:  srcAddr,
			destAddr: destAddr,
			addTime:  time.Now(),
		}
		if runtime.GOOS == "windows" {
			i.Output(&requestData, nil, "")
		} else {
			key := fmt.Sprintf("%s-%s-%d-%d", srcAddr, destAddr, srcPort, destPort)
			locker.Lock()
			i.requests[key] = &requestData
			locker.Unlock()
		}
	}

	if i.isResponse(data) {
		buf := bytes.NewBuffer(data)
		reader := bufio.NewReader(buf)
		response, err := http.ReadResponse(reader, request)
		if err != nil {
			log.Printf("Can't parse response data. %s\n", err.Error())
			return 0, nil
		}

		if i.allowShowResponseBody(response.Header.Get("Content-Type")) {
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
		} else {
			response.Body = nil
		}

		key := fmt.Sprintf("%s-%s-%d-%d", srcAddr, destAddr, srcPort, destPort)
		locker.Lock()
		httpRequestData, _ := i.requests[key]
		delete(i.requests, key)
		locker.Unlock()

		i.Output(httpRequestData, response, i.ReadRawHeader(data))
	}

	return len(data), nil
}

func (i *HttpOutput) Output(requestData *httpRequestData, response *http.Response, rawResponseHeader string) {
	if requestData != nil {
		url := "http://" + requestData.request.Host + requestData.request.RequestURI
		if Setting.Filter != "" && !strings.Contains(url, Setting.Filter) {
			return
		}
	}

	if Setting.Raw {
		i.OutputRAW(requestData, response, rawResponseHeader)
		return
	}

	if requestData != nil && response != nil {
		count++

		fmt.Printf("%-5d %-5s %-5d %-5s %s\n", count, response.StatusCode, response.ContentLength, requestData.request.Method, "http://"+requestData.request.Host+requestData.request.RequestURI)
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()
		}
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println(i.SubString(string(body), 1000))
		}
	} else if requestData != nil {
		count++

		fmt.Printf("%-5d %-5s %-5s %-5s %s\n", count, "-", "-", requestData.request.Method, "http://"+requestData.request.Host+requestData.request.RequestURI)
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()
		}
	} else {
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println(i.SubString(string(body), 1000))
		}
	}
}

func (i *HttpOutput) OutputRAW(requestData *httpRequestData, response *http.Response, rawResponseHeader string) {
	if requestData != nil && response != nil {
		fmt.Println(ColorfulRequest(requestData.header))
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()

			body, _ := ioutil.ReadAll(requestData.request.Body)
			fmt.Println(i.SubString(string(body), 1000))
		}

		fmt.Println(ColorfulRequest(rawResponseHeader))
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println(ColorfulResponse(i.SubString(string(body), 1000)))
		}
	} else if requestData != nil {
		fmt.Println(ColorfulRequest(requestData.header))
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()

			body, _ := ioutil.ReadAll(requestData.request.Body)
			fmt.Println(ColorfulResponse(i.SubString(string(body), 1000)))
		}

	} else {
		fmt.Println(ColorfulRequest(rawResponseHeader))
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			fmt.Println(ColorfulResponse(i.SubString(string(body), 1000)))
		}
	}

	fmt.Println("")
}

func (i *HttpOutput) SubString(text string, maxLen int) string {
	if len(text) > maxLen {
		return text[:maxLen] + "..."
	} else {
		return text
	}
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
	// get ClientHello headshake package
	if srcPort == 443 || destPort == 443 {
		return true
	}

	return false
}

func (i *HttpOutput) allowShowResponseBody(contentType string) bool {
	contentType = strings.ToLower(contentType)
	switch {
	case strings.Contains(contentType, "text/"),
		strings.Contains(contentType, "application/json"):
		return true
	}

	return false
}

func (i *HttpOutput) requestMonitor() {
	for {
		<-time.Tick(2 * time.Second)

		// output timeout request
		timeout := 5 * time.Second
		locker.Lock()
		for key, requestData := range i.requests {
			if requestData.addTime.Add(timeout).Before(time.Now()) {
				i.Output(requestData, nil, "")
				delete(i.requests, key)
			}
		}
		locker.Unlock()
	}
}

func (i *HttpOutput) String() string {
	return "Http Output"
}
