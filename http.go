package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"path/filepath"
	pb "gopkg.in/cheggaaa/pb.v1"
	"io"
)

var (
	client http.Client
)

var (
	acceptRangeHeader   = "Accept-Ranges"
	contentLengthHeader = "Content-Length"
)

type HttpDownloader struct {
	url string
	file string
	par int64
	len int64
}

func NewHttpDownloader(url string, par int) *HttpDownloader {
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := client.Do(req)

	if resp.Header.Get(acceptRangeHeader) == "" {
		log.Fatalf("target url not support range download")
		os.Exit(1)
	}

	//get download range
	clen := resp.Header.Get(contentLengthHeader)
	if clen == "" {
		log.Fatalf("target url not contain Content-Length header")
		os.Exit(1)
	}

	len, err := strconv.ParseInt(clen, 10, 64)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	file := filepath.Base(url)

	return &HttpDownloader{url, file, int64(par), len}
}

//get url
//output to multiple files
func (d *HttpDownloader) Do() ([]string, error) {
	var reterr error
	var ws sync.WaitGroup
	var lock sync.RWMutex
	var ret = make([]string, 0)

	bars := make([]*pb.ProgressBar, 0)
	for j := int64(0); j < d.par; j++ {
		from := (d.len/d.par)*j
		var to int64
		if j < d.par-1 {
			to = (d.len/d.par)*(j+1) - 1
		} else {
			to = d.len
		}
		len := to - from
		bars = append(bars, pb.New64(len).Prefix(fmt.Sprintf("%s-%d", d.file, j)))
	}

	barpool, err := pb.StartPool(bars...)
	if err != nil {
		panic(err)
	}

	var i int64
	for i = 0; i < d.par; i++ {
		ws.Add(1)
		go func(d *HttpDownloader, loop int64) {
			defer ws.Done()
			var bar = bars[loop]

			//range calculate
			var from int64
			var to int64

			from = (d.len/d.par)*loop
			if loop < d.par-1 {
				to = (d.len/d.par)*(loop+1) - 1
			}

			var ranges string
			if to != 0 {
				ranges = fmt.Sprintf("bytes=%d-%d", from, to)
			} else {
				ranges = fmt.Sprintf("bytes=%d-", from)
			}


			//send request
			req, err := http.NewRequest("GET", *url, nil)
			if err != nil {
				reterr = err
				return
			}
			req.Header.Add("Range", ranges)
			if err != nil {
				reterr = err
				return
			}

			//write to file
			resp, err := client.Do(req)
			defer resp.Body.Close()

			if err != nil {
				reterr = err
				return
			}

			fname := filepath.Base(d.url)
			fname = fmt.Sprintf("/tmp/%s-part%d", fname, loop)
			f, err := os.OpenFile(fname, os.O_CREATE | os.O_WRONLY, 0600)

			defer f.Close()
			if err != nil {
				reterr = err
				return
			}

			writer := io.MultiWriter(f, bar)
			io.Copy(writer, resp.Body)

			bar.Finish()

			lock.Lock()
			ret = append(ret, fname)
			lock.Unlock()
		}(d, i)
	}

	ws.Wait()
	barpool.Stop()

	return ret, reterr
}
