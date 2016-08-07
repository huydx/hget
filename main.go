package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"path/filepath"
)

var (
	url    = flag.String("url", "", "url")
	conn   = flag.Int("n", runtime.NumCPU(), "connection")
)

func main() {
	flag.Parse()

	if *url == "" {
		fmt.Errorf("url is required")
		usage()
		os.Exit(1)
	}

	//set up parallel
	i := runtime.GOMAXPROCS(*conn)
	fmt.Printf("using %d parallel \n", i)

	downloader := NewHttpDownloader(*url, i)
	files, err := downloader.Do()
	if err != nil {
		panic(err)
	}

	err = JoinFile(files, filepath.Base(*url))
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
}

func usage() {
	fmt.Println(`hget -url url -n connection`)
}
