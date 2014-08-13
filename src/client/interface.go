package client

import (
	"errors"
	"fmt"
	"scenario"
)

type ClientInterface interface {
	Do(*scenario.Call, bool) (int64, error)
}

var clients = make(map[string]func(string) (ClientInterface, error))

func Register(name string, c func(string) (ClientInterface, error)) {
	clients[name] = c
}

func New(cname string, proxy string) (ClientInterface, error) {
	if client, ok := clients[cname]; ok {
		return client(proxy)
	}

	return nil, errors.New(fmt.Sprintf("No such client %s, please implement it yourself", cname))
}
