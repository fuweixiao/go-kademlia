package kademlia

import (
	"net"
	"strconv"
	"strings"
)

func ParseIpPort(s string) (ipaddr net.IP, port uint16) {
	slices := strings.Split(s, ":")
	if slices[0] == "localhost" {
		ipaddr = net.ParseIP("127.0.0.1")
	} else {
		ipaddr = net.ParseIP(slices[0])
	}
	if tmp, err := strconv.Atoi(slices[1]); err == nil {
		port = uint16(tmp)
	} else {
		port = 0
	}
	return
}
