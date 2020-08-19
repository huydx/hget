package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

var displayProgress = true

func main() {
	var err error
	var proxy string

	conn := flag.Int("n", runtime.NumCPU(), "connection")
	skiptls := flag.Bool("skip-tls", true, "skip verify certificate for https")
	flag.StringVar(&proxy, "proxy", "", "proxy for downloading, ex \n\t-proxy '127.0.0.1:12345' for socks5 proxy\n\t-proxy 'http://proxy.com:8080' for http proxy")

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		Errorln("url is required")
		usage()
		os.Exit(1)
	}

	command := args[0]
	if command == "tasks" {
		if err = TaskPrint(); err != nil {
			Errorf("%v\n", err)
		}
		return
	} else if command == "resume" {
		if len(args) < 2 {
			Errorln("downloading task name is required")
			usage()
			os.Exit(1)
		}

		var task string
		if IsUrl(args[1]) {
			task = TaskFromUrl(args[1])
		} else {
			task = args[1]
		}

		state, err := Resume(task)
		FatalCheck(err)
		Execute(state.Url, state, *conn, *skiptls, proxy)
		return
	} else {
		if ExistDir(FolderOf(command)) {
			Warnf("Downloading task already exist, remove first \n")
			err := os.RemoveAll(FolderOf(command))
			FatalCheck(err)
		}
		Execute(command, nil, *conn, *skiptls, proxy)
	}
}

func Execute(url string, state *State, conn int, skiptls bool, proxy string) {
	//otherwise is hget <URL> command

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

	doneChan := make(chan bool, conn)
	fileChan := make(chan string, conn)
	errorChan := make(chan error, 1)
	stateChan := make(chan Part, 1)
	interruptChan := make(chan bool, conn)

	var downloader *HttpDownloader
	if state == nil {
		downloader = NewHttpDownloader(url, conn, skiptls, proxy)
	} else {
		downloader = &HttpDownloader{url: state.Url, file: filepath.Base(state.Url), par: int64(len(state.Parts)), parts: state.Parts, resumable: true}
	}
	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, stateChan)

	for {
		select {
		case <-signal_chan:
			//send par number of interrupt for each routine
			isInterrupted = true
			for i := 0; i < conn; i++ {
				interruptChan <- true
			}
		case file := <-fileChan:
			files = append(files, file)
		case err := <-errorChan:
			Errorf("%v", err)
			panic(err) //maybe need better style
		case part := <-stateChan:
			parts = append(parts, part)
		case <-doneChan:
			if isInterrupted {
				if downloader.resumable {
					Printf("Interrupted, saving state ... \n")
					s := &State{Url: url, Parts: parts}
					err := s.Save()
					if err != nil {
						Errorf("%v\n", err)
					}
					return
				} else {
					Warnf("Interrupted, but downloading url is not resumable, silently die")
					return
				}
			} else {
				err := JoinFile(files, filepath.Base(url))
				FatalCheck(err)
				err = os.RemoveAll(FolderOf(url))
				FatalCheck(err)
				return
			}
		}
	}
}

func usage() {
	Printf(`Usage:
hget [-n connection] [-skip-tls true] [-proxy proxy_address] URL
hget tasks
hget resume [TaskName]
`)
}
