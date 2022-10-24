package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/imkira/go-task"
)

var displayProgress = true

func main() {
	var err error
	var proxy, filepath, bwLimit string

	conn := flag.Int("n", runtime.NumCPU(), "connection")
	skiptls := flag.Bool("skip-tls", true, "skip verify certificate for https")
	flag.StringVar(&proxy, "proxy", "", "proxy for downloading, ex \n\t-proxy '127.0.0.1:12345' for socks5 proxy\n\t-proxy 'http://proxy.com:8080' for http proxy")
	flag.StringVar(&filepath, "file", "", "filepath that contains links in each line")
	flag.StringVar(&bwLimit, "rate", "", "bandwidth limit to use while downloading, ex\n\t -rate 10kB\n\t-rate 10MiB")

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		if len(filepath) < 2 {
			Errorln("url is required")
			usage()
			os.Exit(1)
		}
		// Creating a SerialGroup.
		g1 := task.NewSerialGroup()
		file, err := os.Open(filepath)
		if err != nil {
			FatalCheck(err)
		}

		defer file.Close()

		reader := bufio.NewReader(file)

		for {
			line, _, err := reader.ReadLine()

			if err == io.EOF {
				break
			}

			g1.AddChild(downloadTask(string(line), nil, *conn, *skiptls, proxy, bwLimit))
		}
		g1.Run(nil)
		return
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
		if IsURL(args[1]) {
			task = TaskFromURL(args[1])
		} else {
			task = args[1]
		}

		state, err := Resume(task)
		FatalCheck(err)
		Execute(state.URL, state, *conn, *skiptls, proxy, bwLimit)
		return
	} else {
		if ExistDir(FolderOf(command)) {
			Warnf("Downloading task already exist, remove first \n")
			err := os.RemoveAll(FolderOf(command))
			FatalCheck(err)
		}
		Execute(command, nil, *conn, *skiptls, proxy, bwLimit)
	}
}

func downloadTask(url string, state *State, conn int, skiptls bool, proxy string, bwLimit string) task.Task {
	run := func(t task.Task, ctx task.Context) {
		Execute(url, state, conn, skiptls, proxy, bwLimit)
	}
	return task.NewTaskWithFunc(run)
}

// Execute configures the HTTPDownloader and uses it to download stuff.
func Execute(url string, state *State, conn int, skiptls bool, proxy string, bwLimit string) {
	//otherwise is hget <URL> command

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
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

	var downloader *HTTPDownloader
	if state == nil {
		downloader = NewHTTPDownloader(url, conn, skiptls, proxy, bwLimit)
	} else {
		downloader = &HTTPDownloader{url: state.URL, file: filepath.Base(state.URL), par: int64(len(state.Parts)), parts: state.Parts, resumable: true}
	}
	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, stateChan)

	for {
		select {
		case <-signalChan:
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
					s := &State{URL: url, Parts: parts}
					if err := s.Save(); err != nil {
						Errorf("%v\n", err)
					}
				} else {
					Warnf("Interrupted, but downloading url is not resumable, silently die")
				}
			} else {
				err := JoinFile(files, filepath.Base(url))
				FatalCheck(err)
				err = os.RemoveAll(FolderOf(url))
				FatalCheck(err)
			}
			return
		}
	}
}

func usage() {
	Printf(`Usage:
hget [-n connection] [-skip-tls true] [-proxy proxy_address] [-file filename] URL
hget tasks
hget resume [TaskName]
`)
}
