package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"path/filepath"
	"io/ioutil"
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

	return &HttpDownloader{url, int64(par), len}
}

//get url
//output to multiple files
func (d *HttpDownloader) Do() ([]string, error) {
	var reterr error
	var ws sync.WaitGroup
	var lock sync.RWMutex
	var ret = make([]string, 0)

	var i int64
	for i = 0; i < d.par; i++ {
		ws.Add(1)
		go func(d *HttpDownloader, loop int64) {
			defer ws.Done()
			fmt.Printf("download part %d start\n", loop)
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
			fmt.Printf("getting ranges %s\n", ranges)

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

			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				reterr = err
				return
			}
			f.Write(bytes)


			lock.Lock()
			ret = append(ret, fname)
			lock.Unlock()
			fmt.Printf("download part %d end\n", loop)

		}(d, i)
	}

	ws.Wait()
	return ret, reterr
}
