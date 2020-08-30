# hget
This project is my personal project to learn golang to build something useful.

![](https://i.gyazo.com/641166ab79e196e35d1a0ef3f9befd80.png)


## Install

```bash
$ go get -d github.com/abzcoding/hget
$ cd $GOPATH/src/github.com/abzcoding/hget
$ make clean install
```

Binary file will be built at ./bin/hget, you can copy to /usr/bin or /usr/local/bin and even `alias wget hget` to replace wget totally :P

## Usage

```bash
hget [-n parallel] [-skip-tls false] [-proxy proxy_server] [-file filename] [URL] # to download url, with n connections, and not skip tls certificate
hget tasks # get interrupted tasks
hget resume [TaskName | URL] # to resume task
hget -proxy "127.0.0.1:12345" URL # to download using socks5 proxy
hget -proxy "http://sample-proxy.com:8080" URL # to download using http proxy
hget -file sample.txt # to download a list of files
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
  -skip-tls
        skip verify certificate for https (default true)
```

To interrupt any on-downloading process, just ctrl-c or ctrl-d at the middle of the download, hget will safely save your data and you will be able to resume later

### Download
![](https://i.gyazo.com/89009c7f02fea8cb4cbf07ee5b75da0a.gif)

### Resume
![](https://i.gyazo.com/caa69808f6377421cb2976f323768dc4.gif)


