package client

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"scenario"
	"strings"
	"time"
)

type HttpClient struct {
	hclient *http.Client
}

func (h *HttpClient) Do(call *scenario.Call, debug bool) (int64, error) {
	// proceeding body type
	req := new(http.Request)

	// escape url
	escapedUrl, _ := url.Parse(call.URL)
	q := escapedUrl.Query()
	escapedUrl.RawQuery = q.Encode()

	// call.BodyType has to hit one of the case
	switch strings.ToUpper(call.BodyType) {
	case "STRING":
		req, _ = http.NewRequest(call.Method, escapedUrl.String(), strings.NewReader(call.Body))
		break
	case "FILE":
		body, err := ioutil.ReadFile(call.Body)
		if err != nil {
			return -1, err
		} else {
			req, _ = http.NewRequest(call.Method, escapedUrl.String(), bytes.NewReader(body))
		}
		break
	default:

		return -1, errors.New("Body type should be either `file` or `string`")
	}

	// dealing with header map
	for k, v := range call.Header {
		req.Header.Set(k, v)
	}

	t1 := time.Now().UnixNano()
	res, err := h.hclient.Do(req)

	response_time := time.Now().UnixNano() - t1
	if res.Body != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return response_time, errors.New(fmt.Sprintf("Request failed for call %s %s", call.Method, call.URL))
	}

	// some output
	if debug {
		// defer res.Body.Close()
		data, _ := ioutil.ReadAll(res.Body)
		log.Println(fmt.Sprintf("REQ : %s %s %s", call.Method, call.URL, call.Body))
		log.Println(fmt.Sprintf("RESP : %s %s", res.Status, string(data)))
	}

	if res.StatusCode >= 400 {
		// log.Println("Got error code --> ", res.Status, "for call ", call.Method, " ", call.URL)
		return response_time, errors.New(fmt.Sprintf("Status --> %d for call %s %s", res.Status, call.Method, call.URL))
	} else {
		// only successful returns
		return response_time, nil
	}

}

func init() {
	Register("http", newHttpClient)
}

func newHttpClient(proxy string) (ClientInterface, error) {
	jar, _ := cookiejar.New(nil)

	if proxy != "nil" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			log.Fatal(err)
		}
		return &HttpClient{
			&http.Client{
				Transport: &http.Transport{
					// DisableKeepAlives:   false,
					// MaxIdleConnsPerHost: 200000,
					Proxy: http.ProxyURL(proxyUrl),
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
				Jar: jar,
			}}, nil
	} else {
		return &HttpClient{
			&http.Client{
				Transport: &http.Transport{
					DisableKeepAlives:   false,
					MaxIdleConnsPerHost: 200000,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
				Jar: jar,
			}}, nil
	}
}
