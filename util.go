package main

import (
	"errors"
	"github.com/mattn/go-isatty"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func FatalCheck(err error) {
	if err != nil {
		Errorf("%v", err)
		panic(err)
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

func FolderOf(url string) string {
	safePath := filepath.Join(os.Getenv("HOME"), dataFolder)
	fullQualifyPath, err := filepath.Abs(filepath.Join(os.Getenv("HOME"), dataFolder, filepath.Base(url)))
	FatalCheck(err)

	//must ensure full qualify path is CHILD of safe path
	//to prevent directory traversal attack
	//using Rel function to get relative between parent and child
	//if relative join base == child, then child path MUST BE real child
	relative, err := filepath.Rel(safePath, fullQualifyPath)
	FatalCheck(err)

	if strings.Contains(relative, "..") {
		FatalCheck(errors.New("you may be a victim of directory traversal path attack\n"))
		return "" //return is redundant be cause in fatal check we have panic, but compiler does not able to check
	} else {
		return fullQualifyPath
	}
}

func TaskFromUrl(url string) string {
	//task is just download file name
	//so we get download file name on url
	filename := filepath.Base(url)
	return filename
}

func IsUrl(s string) bool {
	_, err := url.Parse(s)
	return err == nil
}
