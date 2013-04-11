## hammer.go
Stress test framework in Go   

## Files:
hammer.go - the client hammer tool

## Usage:
1. none session scenario
```
go run hammer.go -rps 100 -profile profile/event_queue_profile.json  
```

2. session scenario
```
go run hammer.go -rps 100 -type ws_session -size 100
```

> try `go run hammer.go -h` to get all cmd parameters
>
> ## To get binary for Hammer:
> **You have to properly compile/update Go for Linux first**
>
> 1. `brew install go --HEAD --cross-compile-common` will resolve all
> 2. `GOOS=linux GOARCH=amd64 CGO_ENABLE=0 go build -o hammer.linux
hammer.go`


## Make it yourself:
* none session scenario

 1. add or modify existed .json files in /profile, following current
    format

* session scenario

 1. create your .go file in /scenario
 2. follow any *_scenario.go file as example

