package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/cxfksword/httpcap/color"
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
	go di.requestMonitor()

	return
}

var request *http.Request
var err error
var locker *sync.Mutex = &sync.Mutex{}
var hasShowHeaderDesc = false

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

		key := i.key(srcAddr, destAddr, srcPort, destPort)
		locker.Lock()
		checkRequestData, found := i.requests[key]
		if found {
			// conflict with prev request
			i.Output(checkRequestData, nil, "")
		}
		i.requests[key] = &requestData
		locker.Unlock()

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

		key := i.key(srcAddr, destAddr, srcPort, destPort)
		locker.Lock()
		httpRequestData, _ := i.requests[key]
		delete(i.requests, key)
		locker.Unlock()

		i.Output(httpRequestData, response, i.ReadRawHeader(data))
	}

	return len(data), nil
}

func (i *HttpOutput) Output(requestData *httpRequestData, response *http.Response, rawResponseHeader string) {
	// filte request
	if requestData != nil {
		url := "http://" + requestData.request.Host + requestData.request.RequestURI
		if Setting.Filter != "" && !strings.Contains(url, Setting.Filter) {
			return
		}
	}

	// raw output mode
	if Setting.Raw {
		i.OutputRAW(requestData, response, rawResponseHeader)
		return
	}

	if strings.TrimSpace(Setting.Format) != "" {
		i.OutputRequestLine(requestData, response, rawResponseHeader)
		return
	}

	if requestData != nil && response != nil {
		i.showHeaderDescription()
		color.Printf("%-24s %-5d %-5d %-5s %s\n",
			color.MethodColor(requestData.request.Method),
			time.Now().Format("2006-01-02 15:04:05"),
			response.StatusCode,
			response.ContentLength,
			requestData.request.Method,
			"http://"+requestData.request.Host+requestData.request.RequestURI)
		i.OutputRequestLine(requestData, response, rawResponseHeader)
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()
		}
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			i.OutputBody(body)
		}
	} else if requestData != nil {
		i.showHeaderDescription()
		color.Printf("%-24s %-5s %-5s %-5s %s\n",
			color.MethodColor(requestData.request.Method),
			time.Now().Format("2006-01-02 15:04:05"),
			"-",
			"-",
			requestData.request.Method,
			"http://"+requestData.request.Host+requestData.request.RequestURI)
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()
		}
	} else {
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			i.OutputBody(body)
		}
	}
}

func (i *HttpOutput) OutputRAW(requestData *httpRequestData, response *http.Response, rawResponseHeader string) {
	if requestData != nil && response != nil {
		color.PrintlnRequest(requestData.header)
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()

			body, _ := ioutil.ReadAll(requestData.request.Body)
			i.OutputBody(body)
		}

		color.Println("Response:", color.Yellow)
		color.PrintlnRequest(rawResponseHeader)
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			i.OutputBody(body)
		}
	} else if requestData != nil {
		color.PrintlnRequest(requestData.header)
		if requestData.request.Body != nil {
			defer requestData.request.Body.Close()

			body, _ := ioutil.ReadAll(requestData.request.Body)
			i.OutputBody(body)
		}

	} else {
		color.Println("Response:", color.Yellow)
		color.PrintlnRequest(rawResponseHeader)
		if response.Body != nil {
			defer response.Body.Close()

			body, _ := ioutil.ReadAll(response.Body)
			i.OutputBody(body)
		}
	}

	fmt.Println("")
}

func (i *HttpOutput) OutputBody(body []byte) {
	content := string(body)
	if Setting.TruncateBodyLength > 0 {
		content = i.SubString(content, Setting.TruncateBodyLength)
	}
	if strings.TrimSpace(content) == "" {
		return
	}
	if i.IsPrintable(content) {
		color.PrintlnResponse(content)
	} else {
		// can't printable char, encode to hex
		color.PrintResponse(hex.EncodeToString([]byte(content)) + " ")
		color.Println("<unprintable characters>", color.Default)
	}
}

func (i *HttpOutput) OutputRequestLine(requestData *httpRequestData, response *http.Response, rawResponseHeader string) {
	requestBody := ""
	if requestData.request != nil && requestData.request.Body != nil {
		bytes, _ := ioutil.ReadAll(requestData.request.Body)
		requestBody = string(bytes)
	}
	responseBody := ""
	if response != nil && response.Body != nil {
		bytes, _ := ioutil.ReadAll(response.Body)
		responseBody = string(bytes)
	}
	variables := map[string]string{
		"%request.time":                     time.Now().Format("2006-01-02 15:04:05"),
		"%dest.ip":                          requestData.destAddr,
		"%dest.port":                        fmt.Sprintf("%d", requestData.destPort),
		"%source.ip":                        requestData.srcAddr,
		"%source.port":                      fmt.Sprintf("%d", requestData.srcPort),
		"%request.method":                   requestData.request.Method,
		"%request.url":                      requestData.request.RequestURI,
		"%request.user-agent":               requestData.request.UserAgent(),
		"%request.host":                     requestData.request.Host,
		"%request.header":                   requestData.header,
		"%request.body":                     requestBody,
		"%request.header.protocol":          requestData.request.Proto,
		"%request.header.accept":            requestData.request.Header.Get("Accept"),
		"%request.header.accept-encoding":   requestData.request.Header.Get("Accept-Encoding"),
		"%request.header.content-length":    fmt.Sprintf("%d", requestData.request.ContentLength),
		"%request.header.cookie":            requestData.request.Header.Get("Cookie"),
		"%request.header.referer":           requestData.request.Referer(),
		"%response.status":                  fmt.Sprintf("%d", response.StatusCode),
		"%response.header":                  rawResponseHeader,
		"%response.body":                    responseBody,
		"%response.header.content-type":     response.Header.Get("Content-Type"),
		"%response.header.content-encoding": response.Header.Get("Content-Encoding"),
		"%response.header.content-length":   fmt.Sprintf("%d", response.ContentLength),
		"%response.header.etag":             response.Header.Get("Etag"),
		"%response.header.cache-control":    response.Header.Get("Cache-Control"),
		"%response.header.connection":       response.Header.Get("Connection"),
		"%response.header.last-modified":    response.Header.Get("Last-Modified"),
		"%response.header.set-cookie":       response.Header.Get("Set-Cookie"),
	}

	// names := map[string]string{
	// 	"%request.time":                     "time",
	// 	"%dest.ip":                          "dest-ip",
	// 	"%dest.port":                        "dest-port",
	// 	"%source.ip":                        "src-ip",
	// 	"%source.port":                      "src-port",
	// 	"%request.method":                   "method",
	// 	"%request.url":                      "url",
	// 	"%request.user-agent":               "useragent",
	// 	"%request.host":                     "host",
	// 	"%request.header":                   "request-header",
	// 	"%request.body":                     "request-body",
	// 	"%request.header.protocol":          "protocol",
	// 	"%request.header.accept":            "accept",
	// 	"%request.header.accept-encoding":   "accept-encoding",
	// 	"%request.header.content-length":    "content-length",
	// 	"%request.header.cookie":            "cookie",
	// 	"%request.header.referer":           "referer",
	// 	"%response.status":                  "status",
	// 	"%response.header":                  "response-header",
	// 	"%response.body":                    "response-body",
	// 	"%response.header.content-type":     "content-type",
	// 	"%response.header.content-encoding": "content-encoding",
	// 	"%response.header.content-length":   "content-length",
	// 	"%response.header.etag":             "etag",
	// 	"%response.header.cache-control":    "cache-control",
	// 	"%response.header.connection":       "connection",
	// 	"%response.header.last-modified":    "last-modified",
	// 	"%response.header.set-cookie":       "set-cookie",
	// }

	var keys = []string{}
	for key, _ := range variables {
		keys = append(keys, key)
	}
	sort.Sort(ByLength(keys))

	line := Setting.Format
	for _, key := range keys {
		if v, found := variables[key]; found {
			line = strings.Replace(line, key, v, -1)
		}
	}

	line = strings.Replace(line, `\n`, "\n", -1)
	line = strings.Replace(line, `\t`, "\t", -1)
	color.Println(line, color.MethodColor(requestData.request.Method))

}

func (i *HttpOutput) showHeaderDescription() {
	if !hasShowHeaderDesc {
		fmt.Printf("%-21s %-5s %-5s %-5s %s\n",
			"time",
			"status",
			"length",
			"method",
			"url")
		hasShowHeaderDesc = true
	}

}

// checks if s is ascii and printable, aka doesn't include tab, backspace, etc.
func (i *HttpOutput) IsPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) && r != rune('\t') && r != rune('\r') && r != rune('\n') {
			return false
		}
	}
	return true
}

func (i *HttpOutput) SubString(text string, maxLen int) string {
	if len(text) > maxLen {
		return text[:maxLen] + " ..."
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
		strings.Contains(contentType, "application/json"),
		strings.Contains(contentType, "application/x-javascript"):
		return true
	}

	return false
}

func (i *HttpOutput) requestMonitor() {
	for {
		<-time.Tick(1 * time.Second)

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

func (i *HttpOutput) key(srcAddr string, destAddr string, srcPort uint16, destPort uint16) string {
	strs := []string{srcAddr, destAddr, fmt.Sprintf("%d", srcPort), fmt.Sprintf("%d", destPort)}
	sort.Strings(strs)
	return strings.Join(strs, "_")
}

func (i *HttpOutput) String() string {
	return "Http Output"
}
