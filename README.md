# hget
This project is my personal project to learn golang to build something useful.

![](https://i.gyazo.com/641166ab79e196e35d1a0ef3f9befd80.png)



**Build Status**: [![Build Status](https://travis-ci.org/huydx/hget.svg?branch=master)](https://travis-ci.org/huydx/hget)

## Install

```
$ go get -d github.com/huydx/hget
$ cd $GOPATH/src/github.com/huydx/hget
$ make clean install
```

Binary file will be built at ./bin/hget, you can copy to /usr/bin or /usr/local/bin and even `alias wget hget` to replace wget totally :P

## Usage

```
hget [Url] [-n parallel] [-skip-tls false] //to download url, with n connections, and not skip tls certificate
hget tasks //get interrupted tasks
hget resume [TaskName | URL] //to resume task
```

To interrupt any on-downloading process, just ctrl-c or ctrl-d at the middle of the download, hget will safely save your data and you will be able to resume later

### Download
![](https://i.gyazo.com/89009c7f02fea8cb4cbf07ee5b75da0a.gif)

### Resume
![](https://i.gyazo.com/caa69808f6377421cb2976f323768dc4.gif)


