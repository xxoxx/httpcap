package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime/debug"
	"strings"

	"github.com/codegangsta/cli"
)

type Flags struct {
	Verbose            bool
	InterfaceName      string
	Port               string
	Format             string
	Raw                bool
	Filter             string
	TruncateBodyLength int
}

var (
	Setting Flags
)

func main() {
	// Don't exit on panic
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(error); !ok {
				fmt.Printf("PANIC: pkg: %v %s \n", r, debug.Stack())
			}
		}
	}()

	app := cli.NewApp()
	app.Name = "httpcap"
	app.Usage = "A simple network analyzer that capture http network traffic."
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "interface, i",
			Value: "",
			Usage: "interface to listen on (e.g. eth0, en1, or 192.168.1.1, 127.0.0.1 etc.)",
		},
		cli.StringFlag{
			Name:  "port, p",
			Value: "",
			Usage: "port to listen on (default listen on all port)",
		},
		cli.BoolFlag{
			Name:  "raw, r",
			Usage: "show raw stream. it is a shortcut for -l %request%response",
		},
		cli.StringFlag{
			Name:  "format, f",
			Value: "",
			Usage: "output format. You can specify the output string format containing reserved keyword that will be replaced with the proper value",
		},
		cli.StringFlag{
			Name:  "keyword, k",
			Value: "",
			Usage: "filte output with the keyword in request url",
		},
		cli.IntFlag{
			Name:  "length, l",
			Value: 500,
			Usage: "the length to truncate response body (0 - no limit)",
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
		Setting.Filter = c.String("keyword")
		Setting.TruncateBodyLength = c.Int("length")
		Setting.Raw = c.Bool("raw")

		if c.Bool("version") {
			cli.ShowVersion(c)
			return
		}

		if c.Bool("help") {
			cli.ShowAppHelp(c)
			return
		}

		startCapture()
	}

	app.Run(os.Args)
}

func ShowAllInterfaces() {
	ifaces, _ := net.Interfaces()

	iplist := ""
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()

		ipV4 := false
		ipAddrs := []string{}
		for _, addr := range addrs {
			var ip net.IP
			if ipnet, ok := addr.(*net.IPNet); ok {
				ip = ipnet.IP
			} else if ipaddr, ok := addr.(*net.IPAddr); ok {
				ip = ipaddr.IP
			}
			if ip != nil && ip.To4() != nil && !ip.IsUnspecified() {
				ipstr := addr.String()
				idx := strings.Index(ipstr, "/")
				if idx >= 0 {
					ipstr = ipstr[:idx]
				}
				ipAddrs = append(ipAddrs, ipstr)
				ipV4 = true
			}
		}
		if !ipV4 {
			continue
		}

		iplist += fmt.Sprintf("%-7d %-40s %s\n", iface.Index, iface.Name, strings.Join(ipAddrs, ", "))
	}

	fmt.Printf("%-7s %-40s %s\n", "index", "interface name", "ip")
	fmt.Print(iplist)
}

func Debug(args ...interface{}) {
	if Setting.Verbose {
		log.Println(args...)
	}
}
