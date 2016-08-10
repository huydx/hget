# hget
![](https://i.gyazo.com/641166ab79e196e35d1a0ef3f9befd80.png)

Rocket fast download accelerator written in golang. Current program working in unix system only.

## Install

```
$ go get -d github.com/huydx/hget
$ cd $GOPATH/src/github.com/huydx/hget
$ make clean install
```

Binary file will be built at ./bin/hget, you can copy to /usr/bin or /usr/local/bin and even `alias wget hget` to replace wget totally :P

## Usage

```
hget url -n parallel -skip-tls false
```

![](https://i.gyazo.com/89009c7f02fea8cb4cbf07ee5b75da0a.gif)


