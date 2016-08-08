package main

import (
	"os"
	"net"
)

func FatalCheck(err error) {
	if err != nil {
		Errorf("%v", err)
		os.Exit(1)
	}
}

func FilterIPV4(ips []net.IP) []string{
	var ret = make([]string, 0)
	for _, ip := range ips {
		if ip.To4() != nil {
			ret = append(ret, ip.String())
		}
	}
	return ret
}
