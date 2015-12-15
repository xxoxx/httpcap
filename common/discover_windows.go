package common

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/go-ps"
)

func DiscoverServices() map[int]Service {
	path := filepath.Join(os.Getenv("SystemRoot"), "System32\\netstat.exe")
	out, err := exec.Command(path, "-ano").Output()
	if err != nil {
		log.Fatal(err)
	}

	pidMap := make(map[int]string)
	process, _ := ps.Processes()
	for _, proc := range process {
		pidMap[proc.Pid()] = filepath.Base(proc.Executable())
	}
	// fmt.Println(pidMap)

	services := make(map[int]Service)
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		reg := regexp.MustCompile("\\s+")
		cols := reg.Split(strings.TrimSpace(line), -1)
		if len(cols) > 4 && cols[0] == "TCP" && cols[3] == "LISTENING" {
			// ignore ipv6
			if strings.HasPrefix(cols[1], "[::]") {
				continue
			}

			arr := strings.Split(cols[1], ":")
			port, _ := strconv.Atoi(arr[1])
			pid, _ := strconv.Atoi(cols[4])
			exec, _ := pidMap[pid]

			switch exec {
			case "redis.exe":
				services[port] = Service{Port: port, Type: Service_Type_Redis, Pid: pid}
			case "memcached.exe":
				services[port] = Service{Port: port, Type: Service_Type_Memcache, Pid: pid}
			case "mongod.exe":
				services[port] = Service{Port: port, Type: Service_Type_Mongodb, Pid: pid}
			case "mysql.exe":
				services[port] = Service{Port: port, Type: Service_Type_Mysql, Pid: pid}
			case "nutcracker.exe":
				services[port] = Service{Port: port, Type: Service_Type_Twemproxy, Pid: pid}
			}

		}
	}

	return services
}
