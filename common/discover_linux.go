package common

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func DiscoverServices() []Service {
	out, err := exec.Command("netstat", "-tnpl | grep -E 'mysqld|redis-server|memcached|mongos|nutcracker' | awk '{print $4,$7}'").Output()
	if err != nil {
		log.Fatal(err)
	}

	services := []Service{}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		reg := regexp.MustCompile("\\s+")
		cols := reg.Split(strings.TrimSpace(line), -1)

		// ignore ipv6
		if strings.HasPrefix(cols[0], ":::") {
			continue
		}

		arr := strings.Split(cols[0], ":")
		port, _ := strconv.Atoi(arr[1])
		arr = strings.Split(cols[1], "/")
		pid, _ := strconv.Atoi(arr[0])
		exec := arr[1]

		switch exec {
		case "redis-server":
			services = append(services, Service{Port: port, Type: Service_Type_Redis, Pid: pid})
		case "memcached":
			services = append(services, Service{Port: port, Type: Service_Type_Memcache, Pid: pid})
		case "mongod":
			services = append(services, Service{Port: port, Type: Service_Type_Mongodb, Pid: pid})
		case "mysqld":
			services = append(services, Service{Port: port, Type: Service_Type_Mysql, Pid: pid})
		case "nutcracker":
			services = append(services, Service{Port: port, Type: Service_Type_Twemproxy, Pid: pid})
		}

	}

	return services
}
