package main

import (
	"flag"
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
	var err error

	flag.Parse()
	if len(os.Args) < 2 {
		Errorln("url is required")
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "tasks" {
		if err = TaskPrint(); err != nil {
			Errorf("%v\n", err)
		}
		return
	} else if command == "resume" {
		if len(os.Args) < 3 {
			Errorln("downloading task name is required")
			usage()
			os.Exit(1)
		}
		Resume(os.Args[2])
		return
	}

	//otherwise is hget <URL> command
	url := command

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//set up parallel

	var files = make([]string, 0)
	var parts = make([]Part, 0)
	var isInterrupted = false

	doneChan := make(chan bool, *conn)
	fileChan := make(chan string, *conn)
	errorChan := make(chan error, 1)
	stateChan := make(chan Part, 1)
	interruptChan := make(chan bool, *conn)

	downloader := NewHttpDownloader(url, *conn, *skiptls)
	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, stateChan)

	for {
		select {
		case <-signal_chan:
			//send par number of interrupt for each routine
			isInterrupted = true
			for i := 0; i < *conn; i++ {
				interruptChan <- true
			}
		case file := <- fileChan:
			files = append(files, file)
		case err := <- errorChan:
			Errorf("%v", err)
			panic(err) //maybe need better style
		case part := <- stateChan:
			parts = append(parts, part)
		case <- doneChan:
			if isInterrupted {
				Printf("Interrupted, saving state ... \n")
				s := &State{Url: url, Parts: parts}
				err := s.Save()
				if err != nil {
					Errorf("%v\n", err)
				}
			} else {
				err = JoinFile(files, filepath.Base(url))
				if err != nil {
					panic(err)
				}
			}
			return
		}
	}

}

func usage() {
	Printf(`Usage:
hget [URL] [-n connection] [-skip-tls true]
hget tasks
hget resume [TaskName]
`)
}
