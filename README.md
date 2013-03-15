hammer.go
=========
Stress test framework in Go   

Files:
======

hammer.go - the client hammer tool   
server.go - a lightweight server just for testing purpose   

To run the performance test:
============================
```shell
GOPATH=$GOPATH:`pwd` go run hammer.go -rps 1   
```

To run test with Oauth:
=======================
```shell
GOPATH=$GOPATH:`pwd` go run hammer.go -rps 1 -auth="oauth"   
```

To run test with Greeauth:
==========================
```shell
GOPATH=$GOPATH:`pwd` go run hammer.go -rps 1 -auth="grees2s"   

GOPATH=$GOPATH:`pwd` go run hammer.go -rps 1 -auth="greec2s"   
```

To enable debug:
================
```shell
GOPATH=$GOPATH:`pwd` go run hammer.go -rps 1 -auth="greec2s" -debug   
```

To build Hammer for Linux:
==========================
You have to properly compile/update Go for Linux first,

To build binary for Linux
```shell
GOOS=linux GOARCH=amd64 GOPATH=$GOPATH:`pwd` CGO_ENABLED=0 go build -o hammer.prod.linux hammer.go
```

To update traffic profile:
==========================

You will have to update the trafficprofiles pkg source (this will be updated with more details, and subject to change)

The file is:
```shell
src/trafficprofiles/trafficprofile.go
```
