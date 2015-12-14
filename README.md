#httpcap
A simple network analyzer that captures http network traffic.

* support windows and linux
* colorful output
* support gzip response body

![screenshot](http://ww3.sinaimg.cn/large/7ce4a9f6gw1esw0oayz8dj20rv08zqb9.jpg)


#Usage

```
NAME:
   httpcap - A simple network analyzer that capture http network traffic.

USAGE:
   httpcap [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   list         show all interfaces
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --interface, -i      interface to listen on (e.g. eth0, en1, or 192.168.1.1, 127.0.0.1 etc.)
   --port, -p           port to listen on (default listen on all port)
   --raw, -r            show raw stream. it is a shortcut for -l %request%response
   --format, -f         output format. You can specify the output string format containing reserved keyword that will be replaced with the proper value
   --keyword, -k        filte output with the keyword in request url
   --length, -l "500"   the length to truncate response body (0 - no limit)
   --verbose, -V        output debug message
   --help, -h           show help
   --version, -v        print the version

```


#Example

```
httpcap -i eth0 -p 80
httpcap -p 80 -k amazon
httpcap -f "%request.time\t%source.ip:%source.port => %dest.ip:%dest.port\thttp://%request.host%request.url\t%response.status"
```

#Compile on windows

1. download [Msys2](https://msys2.github.io/)

2. Download [WinPcap developer pack](https://www.winpcap.org/devel.htm) and extract to ```C:\WpdPack```, make sure the directory is same with ```gopacket/pcap/pcap.go```:

```
#cgo windows CFLAGS: -I C:/WpdPack/Include
```
3. execute command

```
go build main.go capture.go input_raw.go output_http.go reader.go sort.go writer.go
```

>> [Packet Capture, Injection, and Analysis with Gopacket](http://www.devdungeon.com/content/packet-capture-injection-and-analysis-gopacket)

#Run on windows

when run on windows vista or windows 7+, must turn off ```windows firewall``` and ```run with Administrator```.



#Format Variables
