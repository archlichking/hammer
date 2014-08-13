# hammer.go
Stress test framework in Go

## Files
hammer.mac - executable for mac

## Usage 
You don't have to clone the whole repo, just download hammer.mac and .json in src/profile, and run it directly in your OSX by
```
./hammer.mac -r 1 -p src/profile/test_call.json
``` 
Command line arguments
```
./hammer.mac -h
```
Command line arguments explanation

| Arg |                     | Description                                                                                                                                                           |
|-----|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| -D  | debug mode          | if enabled, request/response body will be printed to console log                                                                                                      |
| -P  | proxy               | http proxy server that would be used by hammer for every request                                                                                                      |
| -d  | duration            | duration of the test, by second                                                                                                                            |
| -i  | log output interval | interval for log to be printed in console, by default 6 seconds                                                                                                       |
| -l  | log type            | # not being used right now                                                                                                                                            |
| -m  | load mode           | constant - hammer will send request constantly based on the request # per second. flexible - request will be sent based on average response time                      |
| -p  | json profile        | file location for configuration file .json                                                                                                                            |
| -r  | request per second  | how many request should be sent per second                                                                                                                            |
| -t  | slowness threshold  | when to treat a request as slow request, by millisecond                                                                                                                               |
| -w  | warm up threshold   | if enabled, hammer will warm up the target server by gradually increasing request # per second. Currently built in increment is 1/4 request number by 1/4 warmup time, by second |

## Scenario Configuration
Hammer supports single call, session and call/session hybrid test scenario. 

Single call sample [src/profile/test_call.json](https://gecgithub01.walmart.com/MobileQE/hammer/blob/master/src/profile/test_call.json)
Session sample [src/profile/test_session.json](https://gecgithub01.walmart.com/MobileQE/hammer/blob/master/src/profile/test_session.json)
Hybrid sample [src/profile/test_combination.json](https://gecgithub01.walmart.com/MobileQE/hammer/blob/master/src/profile/test_combination.json)
> more samples could be found in [src/profile/](https://gecgithub01.walmart.com/MobileQE/hammer/blob/master/src/profile)

### Config specification
Basic json structure needs to follow
```
{
    "Scenarios": [{
        "Weight": int,
        "Type": string(call | session),
        "Groups": [{
            "Weight": int,
            "Calls": [{
                "URL": string,
                "Method": string(get | post | put | ...),
                "Type": string(http),
                "BodyType": string(string | file),
                "Body": string | null
                "Header": json object | null
            }]
        }]
    }]
}
``` 
> 1. If BodyType = file, hammer will use the file declared in Body as request body
> 2. Body can be null

Randomization support in config .json

Randomization is supported in URL and Body.
 
1.  You can declare collections of items under Vairables, and refer each collection by its name. Hammer will randomly pick item from a collection during running. 

```
{
    "Variables": {
        "QUERY": ["Pad", "games", "Shoes", "Tea", "Milk", "step stool"],
        "STORE": ["1", "2"]
    },
    "Scenarios": [{
        "Weight": 100,
        "Type": "call",
        "Groups": [{
            "Weight": 100,
            "Calls": [{
                "URL": "http://search.walmart.com/search?query=${QUERY}&store=${STORE}",
                "Method": "GET",
                "Type": "HTTP",
                "BodyType": "STRING",
                "Body": null
            }]
        }]
    }]
}
```

2.  Built in randomized method
 	*   _random_range_float_(int, int) will generate a random float in [int, int)
 	*   _random_range_int_(int, int) will generate a random integer in [int, int)

> More randomizations are being added

## Compile it yourself:
To compile hammer.go, you have to install golang first. Benefit for this is you can compile hammer.go for any platform that golang supports ([full os list](http://golang.org/doc/install/source)). Easiest way to install golang itself is via
```
brew install go --HEAD --cross-compile-common
```

When everything is set, simply run 
```
GOOS=linux GOARCH=amd64 CGO_ENABLE=0 go build -o hammer.linux hammer.go
```

