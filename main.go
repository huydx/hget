package main

import (
	"flag"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"runtime"
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

	//set up parallel
	downloader := NewHttpDownloader(url, *conn, *skiptls)
	files, err := downloader.Do()
	if err != nil {
		panic(err)
	}

	err = JoinFile(files, filepath.Base(url))
	if err != nil {
		panic(err)
	}
}

func usage() {
	Printf("%s: hget [URL] [-n connection]", color.CyanString("Usage"))
}
