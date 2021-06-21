[![Build Status](https://travis-ci.com/abzcoding/hget.svg?branch=master)](https://travis-ci.com/abzcoding/hget)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/abzcoding/hget/badges/quality-score.png?b=master)](https://scrutinizer-ci.com/g/abzcoding/hget/?branch=master)
# hget
![](https://i.gyazo.com/641166ab79e196e35d1a0ef3f9befd80.png)

### Features
- Fast (multithreading & stuff)
- Ability to interrupt/resume (task mangement)
- Support for proxies( socks5 or http)
- Bandwidth limiting
- You can give it a file that contains list of urls to download

### Install

```bash
$ go get -d github.com/abzcoding/hget
$ cd $GOPATH/src/github.com/abzcoding/hget
$ make clean install
```

Binary file will be built at ./bin/hget, you can copy to /usr/bin or /usr/local/bin and even `alias wget hget` to replace wget totally :P

### Usage

```bash
hget [-n parallel] [-skip-tls false] [-rate bwRate] [-proxy proxy_server] [-file filename] [URL] # to download url, with n connections, and not skip tls certificate
hget tasks # get interrupted tasks
hget resume [TaskName | URL] # to resume task
hget -proxy "127.0.0.1:12345" URL # to download using socks5 proxy
hget -proxy "http://sample-proxy.com:8080" URL # to download using http proxy
hget -file sample.txt # to download a list of files
hget -n 4 -rate 100KB URL # to download using 4 threads & limited to 100Kb per second
```

### Help
```
[I] ➜ hget -h
Usage of hget:
  -file string
        filepath that contains links in each line
  -n int
        connection (default 16)
  -proxy string
        proxy for downloading, ex
                -proxy '127.0.0.1:12345' for socks5 proxy
                -proxy 'http://proxy.com:8080' for http proxy
  -rate string
        bandwidth limit to use while downloading, ex
                -rate 10kB
                -rate 10MiB
  -skip-tls
        skip verify certificate for https (default true)
```

To interrupt any on-downloading process, just ctrl-c or ctrl-d at the middle of the download, hget will safely save your data and you will be able to resume later

### Download
![](https://i.gyazo.com/89009c7f02fea8cb4cbf07ee5b75da0a.gif)

### Resume
![](https://i.gyazo.com/caa69808f6377421cb2976f323768dc4.gif)


