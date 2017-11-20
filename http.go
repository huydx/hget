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
	parts     []Part
	resumable bool
}

func NewHttpDownloader(url string, par int, skipTls bool) *HttpDownloader {
	var resumable = true

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

	return ret
}

func partCalculate(par int64, len int64, url string) []Part {
	// Pre-allocate, perf tunning
	ret := make([]Part, par)
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

		// Padding 0 before path name as filename will be sorted as string
		fname := fmt.Sprintf("%s.part%06d", file, j)
		path := filepath.Join(folder, fname) // ~/.hget/download-file-name/part-name
		ret[j] = Part{Index: j, Url: url, Path: path, RangeFrom: from, RangeTo: to}
	}

	return ret
}

func (d *HttpDownloader) Do(doneChan chan bool, fileChan chan string, errorChan chan error, interruptChan chan bool, stateSaveChan chan Part) {
	var ws sync.WaitGroup
	var bars []*pb.ProgressBar
	var barpool *pb.Pool
	var err error

	for _, p := range d.parts {

		if p.RangeTo <= p.RangeFrom {
			fileChan <- p.Path
			stateSaveChan <- Part{
				Index:     p.Index,
				Url:       d.url,
				Path:      p.Path,
				RangeFrom: p.RangeFrom,
				RangeTo:   p.RangeTo,
			}

			continue
		}

		var bar *pb.ProgressBar

		if DisplayProgressBar() {
			bar = pb.New64(p.RangeTo - p.RangeFrom).SetUnits(pb.U_BYTES).Prefix(color.YellowString(fmt.Sprintf("%s-%d", d.file, p.Index)))
			bars = append(bars, bar)
		}

		ws.Add(1)
		go func(d *HttpDownloader, bar *pb.ProgressBar, part Part) {
			defer ws.Done()

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
			resp, err := client.Do(req)
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

			current := int64(0)
			finishDownloadChan := make(chan bool)

			go func() {
				written, _ := io.Copy(writer, resp.Body)
				current += written
				fileChan <- part.Path
				finishDownloadChan <- true
			}()

			select {
			case <-interruptChan:
				// interrupt download by forcefully close the input stream
				resp.Body.Close()
				<-finishDownloadChan
			case <-finishDownloadChan:
			}

			stateSaveChan <- Part{
				Index:     part.Index,
				Url:       d.url,
				Path:      part.Path,
				RangeFrom: current + part.RangeFrom,
				RangeTo:   part.RangeTo,
			}

			if DisplayProgressBar() {
				bar.Update()
				bar.Finish()
			}
		}(d, bar, p)
	}

	barpool, err = pb.StartPool(bars...)
	FatalCheck(err)

	ws.Wait()
	doneChan <- true
	barpool.Stop()
}
