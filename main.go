package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/codegangsta/cli"
)

type Flags struct {
	Verbose       bool
	InterfaceName string
	Port          string
	Format        string
	Raw           bool
}

var Setting Flags

func main() {
	// Don't exit on panic
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(error); !ok {
				fmt.Printf("PANIC: pkg: %v %s \n", r, debug.Stack())
			}
		}
	}()

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
		})
		http.ListenAndServe(":80", nil)
	}()

	app := cli.NewApp()
	app.Name = "httpsf"
	app.Usage = "A simple network analyzer that captures http network traffic."
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "interface, i",
			Value: "",
			Usage: "interface to listen on (e.g. eth0, en1, or 192.168.1.1, 127.0.0.1 etc.)",
		},
		cli.StringFlag{
			Name:  "port, p",
			Value: "80",
			Usage: "port to listen on (default listen on all port)",
		},
		cli.BoolFlag{
			Name:  "raw, r",
			Usage: "show raw stream. it is a shortcut for -l %request%response",
		},
		cli.StringFlag{
			Name:  "format, f",
			Value: "",
			Usage: "log format. You can specify the output string format containing reserved keyword that will be replaced with the proper value",
		},
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "output debug message",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "show all interfaces",
			Action: func(c *cli.Context) {
				ShowAllInterfaces()
			},
		}}
	app.Action = func(c *cli.Context) {
		Setting.Verbose = c.Bool("verbose")
		Setting.InterfaceName = c.String("interface")
		Setting.Port = c.String("port")
		Setting.Format = c.String("format")
		Setting.Raw = c.Bool("raw")
		startCapture()
	}

	app.Run(os.Args)
}

func ShowAllInterfaces() {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()

		ipV4 := false
		ipAddrs := []string{}
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPAddr); ok && !ip.IP.IsUnspecified() {
				ipAddrs = append(ipAddrs, addr.String())
				ipV4 = true
			}
		}
		if !ipV4 {
			continue
		}

		fmt.Printf("%d %s %s\n", iface.Index, iface.Name, strings.Join(ipAddrs, ", "))
	}
}

func GetFirstInterface() (name string, ip string) {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()

		ipV4 := false
		ipAddrs := []string{}
		for _, addr := range addrs {
			if ip, ok := addr.(*net.IPAddr); ok && !ip.IP.IsUnspecified() {
				ipAddrs = append(ipAddrs, addr.String())
				ipV4 = true
			}
		}
		if !ipV4 {
			continue
		}

		return iface.Name, ipAddrs[0]
	}

	return "localhost", "127.0.0.1"
}

func GetIp(iface *net.Interface) string {
	addrs, _ := iface.Addrs()

	ipAddrs := []string{}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPAddr); ok && !ip.IP.IsUnspecified() {
			ipAddrs = append(ipAddrs, addr.String())
		}
	}

	if len(ipAddrs) > 0 {
		return ipAddrs[0]
	} else {
		return ""
	}
}

func Debug(args ...interface{}) {
	if Setting.Verbose {
		log.Println(args...)
	}
}
