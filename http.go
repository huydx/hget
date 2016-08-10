package main

import (
	"crypto/tls"
	"fmt"
	"github.com/fatih/color"
	pb "gopkg.in/cheggaaa/pb.v1"
	"io"
	"net"
	"net/http"
	stdurl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	url       string
	file      string
	par       int64
	len       int64
	ips       []string
	skipTls   bool
}

func NewHttpDownloader(url string, par int, skipTls bool) *HttpDownloader {
	parsed, err := stdurl.Parse(url)
	FatalCheck(err)

	ips, err := net.LookupIP(parsed.Host)
	FatalCheck(err)

	ipstr := FilterIPV4(ips)
	Printf("Resolve ip: %s\n", strings.Join(ipstr, " | "))

	req, err := http.NewRequest("GET", url, nil)
	FatalCheck(err)

	resp, err := client.Do(req)
	FatalCheck(err)

	if resp.Header.Get(acceptRangeHeader) == "" {
		Warnf("Target url is not supported range download\n")
		//fallback to par = 1
		par = 1
	}

	//get download range
	clen := resp.Header.Get(contentLengthHeader)
	if clen == "" {
		Warnf("Target url not contain Content-Length header\n")
		clen = "1" //set 1 because of progress bar not accept 0 length
	}

	len, err := strconv.ParseInt(clen, 10, 64)
	FatalCheck(err)

	sizeInMb := float64(len) / (1000 * 1000)

	if clen == "1" {
		Printf("Dowload size: not specified\n")
	} else if sizeInMb < 1000 {
		Printf("Download target size: %.1f MB\n", sizeInMb)
	} else {
		Printf("Download target size: %.1f GB\n", sizeInMb/1000)
	}

	file := filepath.Base(url)

	return &HttpDownloader{url, file, int64(par), len, ipstr, skipTls}
}

func (d *HttpDownloader) Do(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool, stateSaveChan chan Part) {
	var ws sync.WaitGroup

	bars := make([]*pb.ProgressBar, 0)
	for j := int64(0); j < d.par; j++ {
		from := (d.len / d.par) * j
		var to int64
		if j < d.par-1 {
			to = (d.len/d.par)*(j+1) - 1
		} else {
			to = d.len
		}
		len := to - from
		newbar := pb.New64(len).SetUnits(pb.U_BYTES).Prefix(color.YellowString(fmt.Sprintf("%s-%d", d.file, j)))
		if len <= 1 {
			newbar.ShowPercent = false
		}
		bars = append(bars, newbar)
	}

	barpool, err := pb.StartPool(bars...)
	FatalCheck(err)

	var i int64
	for i = 0; i < d.par; i++ {
		ws.Add(1)
		go func(d *HttpDownloader, loop int64) {
			defer ws.Done()
			var bar = bars[loop]

			//range calculate
			var from int64
			var to int64

			from = (d.len / d.par) * loop
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
			resp, err := client.Do(req)
			if err != nil {
				errorChan <- err
				return
			}
			defer resp.Body.Close()


			fname := filepath.Base(d.url)
			fname = fmt.Sprintf("/tmp/%s-part%d", fname, loop)
			f, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0600)

			defer f.Close()
			if err != nil {
				errorChan <- err
				return
			}

			writer := io.MultiWriter(f, bar)

			//make copy interruptable by copy 100 bytes each loop
			current := int64(0)
			for {
				select {
				case <- interruptChan:
					stateSaveChan <- Part{Url: d.url, Path: fname, RangeFrom: current+from, RangeTo: to}
					return
				default:
					written, err := io.CopyN(writer, resp.Body, 100)
					current += written
					if err != nil {
						if err != io.EOF {
							errorChan <- err
						}
						bar.Finish()
						fileChan <- fname
						return
					}
				}
			}
		}(d, i)
	}

	ws.Wait()
	doneChan <- true
	barpool.Stop()
}
