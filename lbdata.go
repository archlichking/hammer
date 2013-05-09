package main

import (
	"flag"
	"fmt"
	"greeauth"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

var (
	lower int
	upper int

	gree_client = new(greeauth.Client)

	client *http.Client
)

func init() {
	flag.IntVar(&lower, "start", 1, "")
	flag.IntVar(&upper, "end", 10000, "")
}

func remove(i int) {
	_m := "DELETE"
	_u := fmt.Sprintf("http://ldb-us.gree-apps.net/v2/qa-test/leaderboards/leaderboardida%d", i)
	_b := ""

	req, err := http.NewRequest(_m, _u, strings.NewReader(_b))
	if err != nil {
		log.Println(err)
		return
	}

	_signature, _timestamp, _ := gree_client.SignS2SRequest(_m, _u, _b)

	req.Header.Add("Authorization", "S2S"+" realm=\"modern-war\""+
		", signature=\""+_signature+"\", timestamp=\""+_timestamp+"\"")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	data, _ := ioutil.ReadAll(res.Body)
	log.Println("---->", req, res, string(data))

	req.Body.Close()
	res.Body.Close()
}

func add(i int) {
	_m := "POST"
	_u := "http://ldb-us.gree-apps.net/v2/qa-test/leaderboards"
	_b := fmt.Sprintf("{\"id\":\"leaderboardida%d\",\"name\":\"leaderboardnamea%d\",\"expiresAt\":\"2013-06-03T21:46:48+00:00\"}", i, i)

	req, err := http.NewRequest(_m, _u, strings.NewReader(_b))
	if err != nil {
		log.Println(err)
		return
	}
	_signature, _timestamp, _ := gree_client.SignS2SRequest(_m, _u, _b)

	req.Header.Add("Authorization", "S2S"+" realm=\"modern-war\""+
		", signature=\""+_signature+"\", timestamp=\""+_timestamp+"\"")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("---->>", req, _b, res.Body)

	defer req.Body.Close()
	defer res.Body.Close()
}

func main() {
	flag.Parse()

	client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 200000,
		},
	}
	gree_client.S2S_Secret = "923ee6aa9d1096e5a366b3a987e961a7"

	var wg sync.WaitGroup

	for i := 1; i < upper; i++ {
		wg.Add(1)
		go func(j int) {
			remove(j)
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 1; i < lower; i++ {
		wg.Add(1)
		go func(j int) {
			add(j)
			wg.Done()
		}(i)
	}
	wg.Wait()

}
