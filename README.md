#httpcap
A simple network analyzer that captures http network traffic.

* support windows and linux
* colorful output
* supprot show gzip response body

#description

```
NAME:
   httpcap - A simple network analyzer that captures http network traffic.

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
   --length, -l "500"   the length to truncate response body (default 500, 0 - no limit)
   --verbose, -V        output debug message
   --help, -h           show help
   --version, -v        print the version
```



#example

```
httpcap -i eth0 -p 80
httpcap -p 80 -k amazon
httpcap -f  "%request.time\t%source.ip:%source.port => %dest.ip:%dest.port\thttp://%request.host%request.url\t%response.status"
```


#options


#format variables