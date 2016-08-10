package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"runtime"
	"os/signal"
	"syscall"
)

var (
	conn    = flag.Int("n", runtime.NumCPU(), "connection")
	skiptls = flag.Bool("skip-tls", true, "skip verify certificate for https")
)

func main() {
	flag.Parse()
	if len(os.Args) < 2 {
		Errorln("url is required")
		usage()
		os.Exit(1)
	}

	url := os.Args[1]

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//set up parallel

	var files = make([]string, 0)

	doneChan := make(chan bool, *conn)
	fileChan := make(chan string, *conn)
	errorChan := make(chan error, 1)

	downloader := NewHttpDownloader(url, *conn, *skiptls)
	go downloader.Do(doneChan, fileChan, errorChan)

	for {
		select {
		case <-signal_chan:
			//process signal later
		case file := <- fileChan:
			files = append(files, file)
		case err := <- errorChan:
			fmt.Println(err)
			panic(err) //maybe need better style
		case <- doneChan:
			err := JoinFile(files, filepath.Base(url))
			if err != nil {
				panic(err)
			}
			return
		}
	}

}

func usage() {
	Printf("%s: hget [URL] [-n connection] [-skip-tls true]", color.CyanString("Usage"))
}
