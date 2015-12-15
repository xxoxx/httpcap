package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/codegangsta/cli"
	"github.com/cxfksword/httpcap/common"
	"github.com/cxfksword/httpcap/config"
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
			Name:  "body, b",
			Value: 0,
			Usage: "the length to truncate response body (0 - not show body, -1 - show all body)",
		},
		cli.BoolFlag{
			Name:  "verbose, vv",
			Usage: "output debug message",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "show all interfaces",
			Action: func(c *cli.Context) {
				common.ShowAllInterfaces()
			},
		}}
	app.Action = func(c *cli.Context) {
		config.Setting.Verbose = c.Bool("verbose")
		config.Setting.InterfaceName = c.String("interface")
		config.Setting.Port = c.String("port")
		config.Setting.Format = c.String("format")
		config.Setting.Filter = c.String("keyword")
		config.Setting.TruncateBodyLength = c.Int("body")
		config.Setting.Raw = c.Bool("raw")

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
