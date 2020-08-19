package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	stdurl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/net/proxy"
	pb "gopkg.in/cheggaaa/pb.v1"
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{Transport: tr}
)

var (
	acceptRangeHeader   = "Accept-Ranges"
	contentLengthHeader = "Content-Length"
)

type HttpDownloader struct {
	proxy     string
	url       string
	file      string
	par       int64
	len       int64
	ips       []string
	skipTls   bool
	parts     []Part
	resumable bool
}

func NewHttpDownloader(url string, par int, skipTls bool, socks5_proxy string) *HttpDownloader {
	var resumable = true

	// setup a http client
	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}

	// set our socks5 as the dialer
	if len(socks5_proxy) > 0 {
		// create a socks5 dialer
		dialer, err := proxy.SOCKS5("tcp", socks5_proxy, nil, proxy.Direct)
		if err != nil {
			fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
			os.Exit(1)
		}
		httpTransport.Dial = dialer.Dial
	}

	parsed, err := stdurl.Parse(url)
	FatalCheck(err)

	ips, err := net.LookupIP(parsed.Host)
	FatalCheck(err)

	ipstr := FilterIPV4(ips)
	Printf("Resolve ip: %s\n", strings.Join(ipstr, " | "))

	req, err := http.NewRequest("GET", url, nil)
	FatalCheck(err)

	resp, err := httpClient.Do(req)
	FatalCheck(err)

	if resp.Header.Get(acceptRangeHeader) == "" {
		Printf("Target url is not supported range download, fallback to parallel 1\n")
		//fallback to par = 1
		par = 1
	}

	//get download range
	clen := resp.Header.Get(contentLengthHeader)
	if clen == "" {
		Printf("Target url not contain Content-Length header, fallback to parallel 1\n")
		clen = "1" //set 1 because of progress bar not accept 0 length
		par = 1
		resumable = false
	}

	Printf("Start download with %d connections \n", par)

	len, err := strconv.ParseInt(clen, 10, 64)
	FatalCheck(err)

	sizeInMb := float64(len) / (1024 * 1024)

	if clen == "1" {
		Printf("Download size: not specified\n")
	} else if sizeInMb < 1024 {
		Printf("Download target size: %.1f MB\n", sizeInMb)
	} else {
		Printf("Download target size: %.1f GB\n", sizeInMb/1024)
	}

	file := filepath.Base(url)
	ret := new(HttpDownloader)
	ret.url = url
	ret.file = file
	ret.par = int64(par)
	ret.len = len
	ret.ips = ipstr
	ret.skipTls = skipTls
	ret.parts = partCalculate(int64(par), len, url)
	ret.resumable = resumable
	ret.proxy = socks5_proxy

	return ret
}

func partCalculate(par int64, len int64, url string) []Part {
	ret := make([]Part, 0)
	for j := int64(0); j < par; j++ {
		from := (len / par) * j
		var to int64
		if j < par-1 {
			to = (len/par)*(j+1) - 1
		} else {
			to = len
		}

		file := filepath.Base(url)
		folder := FolderOf(url)
		if err := MkdirIfNotExist(folder); err != nil {
			Errorf("%v", err)
			os.Exit(1)
		}

		fname := fmt.Sprintf("%s.part%d", file, j)
		path := filepath.Join(folder, fname) // ~/.hget/download-file-name/part-name
		ret = append(ret, Part{Url: url, Path: path, RangeFrom: from, RangeTo: to})
	}
	return ret
}

func (d *HttpDownloader) Do(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool, stateSaveChan chan Part) {
	var ws sync.WaitGroup
	var bars []*pb.ProgressBar
	var barpool *pb.Pool
	var err error

	if DisplayProgressBar() {
		bars = make([]*pb.ProgressBar, 0)
		for i, part := range d.parts {
			newbar := pb.New64(part.RangeTo - part.RangeFrom).SetUnits(pb.U_BYTES).Prefix(color.YellowString(fmt.Sprintf("%s-%d", d.file, i)))
			bars = append(bars, newbar)
		}
		barpool, err = pb.StartPool(bars...)
		FatalCheck(err)
	}

	for i, p := range d.parts {
		ws.Add(1)
		go func(d *HttpDownloader, loop int64, part Part) {
			// setup a http client
			httpTransport := &http.Transport{}
			httpClient := &http.Client{Transport: httpTransport}

			// set our socks5 as the dialer
			if len(d.proxy) > 0 {
				// create a socks5 dialer
				dialer, err := proxy.SOCKS5("tcp", d.proxy, nil, proxy.Direct)
				if err != nil {
					fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
					os.Exit(1)
				}
				httpTransport.Dial = dialer.Dial
			}
			defer ws.Done()
			var bar *pb.ProgressBar

			if DisplayProgressBar() {
				bar = bars[loop]
			}

			var ranges string
			if part.RangeTo != d.len {
				ranges = fmt.Sprintf("bytes=%d-%d", part.RangeFrom, part.RangeTo)
			} else {
				ranges = fmt.Sprintf("bytes=%d-", part.RangeFrom) //get all
			}

			//send request
			req, err := http.NewRequest("GET", d.url, nil)
			if err != nil {
				errorChan <- err
				return
			}

			if d.par > 1 { //support range download just in case parallel factor is over 1
				req.Header.Add("Range", ranges)
				if err != nil {
					errorChan <- err
					return
				}
			}

			//write to file
			resp, err := httpClient.Do(req)
			if err != nil {
				errorChan <- err
				return
			}
			defer resp.Body.Close()
			f, err := os.OpenFile(part.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)

			defer f.Close()
			if err != nil {
				Errorf("%v\n", err)
				errorChan <- err
				return
			}

			var writer io.Writer
			if DisplayProgressBar() {
				writer = io.MultiWriter(f, bar)
			} else {
				writer = io.MultiWriter(f)
			}

			//make copy interruptable by copy 100 bytes each loop
			current := int64(0)
			for {
				select {
				case <-interruptChan:
					stateSaveChan <- Part{Url: d.url, Path: part.Path, RangeFrom: current + part.RangeFrom, RangeTo: part.RangeTo}
					return
				default:
					written, err := io.CopyN(writer, resp.Body, 100)
					current += written
					if err != nil {
						if err != io.EOF {
							errorChan <- err
						}
						bar.Finish()
						fileChan <- part.Path
						return
					}
				}
			}
		}(d, int64(i), p)
	}

	ws.Wait()
	doneChan <- true
	barpool.Stop()
}
