package client

import (
	"errors"
	"fmt"
	"net"
	"scenario"
)

type SocketClient struct {
}

func (s *SocketClient) Do(call *scenario.Call, debug bool) (int64, error) {
	_, err := net.Dial("tcp", fmt.Sprintf("%s:%d", call.Host, call.Port))
	if err != nil {
		return -1, errors.New(fmt.Sprintf("dial error %s:%d", call.Host, call.Port))
	}

	return 0, nil
}

func init() {
	Register("socket", newSocketClient)
}

func newSocketClient(proxy string) (ClientInterface, error) {
	return &SocketClient{}, nil
}
