package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func logg(rw http.ResponseWriter, req *http.Request) {
	var f string
	if f = req.URL.Query().Get("file"); f == "" {
		f = "filelog.txt"
	}
	tts := strings.Split(f, "_")
	bs, _ := ioutil.ReadFile(f)

	str := "[['Time', 'Request/s', 'Response/s', 'Errors', 'Slow Response/s', 'Pending Request', 'Average Response/ms'],"
	str += string(bs)

	str = str[:len(str)-1] + "]"

	// "Performance Log [type:%s rps:%d size:%d slow:%d]"

	p := struct {
		Title string
		Data  string
	}{
		Title: fmt.Sprintf(
			"Performance Log [type:%s rps:%s session size:%s slow thresold:%sms]",
			strings.Join(tts[:len(tts)-3], "_"),
			tts[len(tts)-3],
			tts[len(tts)-2],
			tts[len(tts)-1],
		),
		Data: str,
	}
	t, _ := template.ParseFiles("log.tpl")
	t.Execute(rw, p)
}

func health(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Content-Length", strconv.Itoa(len("health")))
	rw.WriteHeader(200)
	rw.Write([]byte("health"))
}

func main() {
	http.HandleFunc("/log", logg)
	http.HandleFunc("/health", health)
	http.ListenAndServe(":9999", nil)
}
