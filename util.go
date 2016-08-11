package main

import (
	"net"
	"os"
	"github.com/mattn/go-isatty"
)

func FatalCheck(err error) {
	if err != nil {
		Errorf("%v", err)
		os.Exit(1)
	}
}

func FilterIPV4(ips []net.IP) []string {
	var ret = make([]string, 0)
	for _, ip := range ips {
		if ip.To4() != nil {
			ret = append(ret, ip.String())
		}
	}
	return ret
}

func MkdirIfNotExist(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		if err = os.MkdirAll(folder, 0700); err != nil {
			return err
		}
	}
	return nil
}

func ExistDir(folder string) bool {
	_, err := os.Stat(folder)
	return err == nil
}

func DisplayProgressBar() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && displayProgress
}
