package scenario

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"git.gree-dev.net/growth-revenue/uplink/stream/message"
	"git.gree-dev.net/growth-revenue/uplink/stream/socket"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

type UplinkScenario struct {
	SessionScenario
	SessionAmount int

	_gClients             []*Client
	_totalSendCount       int64
	_totalSendReceiveTime int64
}

type Client struct {
	id     string
	hub    string
	socket socket.Interface
	closed bool

	Token  string
	Stream struct {
		Hostname string
		Port     int
	}
}

type SubData struct {
	Type    string `json:"type"`
	Payload int64  `json:"payload"`
}

func (self *Client) Close() {
	if self.closed {
		return
	}

	self.closed = true

	if self.socket != nil {
		self.socket.Close()
	}
}

func (self *Client) Connect(ups *UplinkScenario) {
	// connected as game client
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", self.Stream.Hostname, self.Stream.Port+30000))
	if err != nil {
		log.Println(err)
		self.Close()
		return
	}
	self.socket = socket.NewSocket(conn)

	// Subscribe on stream.
	err = self.socket.SendJSON(&message.Outgoing{
		Type: "subscribe",
		Payload: struct {
			ID    string `json:"id"`
			Hub   string `json:"hub"`
			Token string `json:"token"`
		}{
			ID:    self.id,
			Hub:   self.hub,
			Token: self.Token,
		},
	})
	if err != nil {
		log.Println(err)
		self.Close()
		return
	}
	ssd := new(SubData)
	go func() {
		var im []byte

		for {
			if err := self.socket.Receive(&im); err != nil {
				log.Fatalln(err)
				break
			}
			r := bytes.NewReader(im)
			json.NewDecoder(r).Decode(&ssd)
			if ssd.Payload != 0 {
				atomic.AddInt64(&ups._totalSendCount, 1)
				gap := time.Now().UnixNano() - ssd.Payload
				atomic.AddInt64(&ups._totalSendReceiveTime, gap)
			}
		}

		self.Close()
		// log.Println(fmt.Sprintf("%2.2f%s", (float64(ups._totalReceiveTime) / float64(ups._totalSend)), "s"))

	}()
}

func (ss *UplinkScenario) InitFromCode() {
	ss._sessions = make([]*Session, ss.SessionAmount)

	// _HOST := "http://172.30.52.157:8080"
	_HOST := "http://192.168.1.100:8080"
	_HUB := "war-of-nations"
	ss._gClients = make([]*Client, ss.SessionAmount)
	for i, _ := range ss._gClients {
		ss._gClients[i] = new(Client)
		ss._gClients[i].id = strconv.Itoa(i + 1)
		ss._gClients[i].hub = _HUB
	}

	for i := 0; i < ss.SessionAmount; i++ {
		ss.addSession([]GenSession{
			GenSession(func() (float32, GenCall, GenCallBack) {
				k := i
				seq := strconv.Itoa(i + 1)
				return 0,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {

						return "POST", "REST",
							_HOST + "/v1/" + _HUB + "/subscribers/" + seq,
							"{\"channels\": [\"/cc/1\", \"/cc/2\"]}"
					}),
					GenCallBack(func(se *Session, st int, storage []byte) {
						se.InternalLock.Lock()
						defer se.InternalLock.Unlock()
						se.State += st
						se.StepLock <- se.State
						// do the game client connection here
						r := bytes.NewReader(storage)

						json.NewDecoder(r).Decode(ss._gClients[k])
						ss._gClients[k].Connect(ss)
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				seq := strconv.Itoa(i + 1)
				return 50,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						t1 := strconv.FormatInt(time.Now().UnixNano(), 10)
						return "POST", "REST",
							_HOST + "/v1/" + _HUB + "/subscribers/" + seq + "/send",
							"{\"type\":\"subscribed\",\"payload\":" + t1 + "}"
					}),
					nil
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				seq := strconv.Itoa(i + 1)
				return 50,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						t1 := strconv.FormatInt(time.Now().UnixNano(), 10)
						return "POST", "REST",
							_HOST + "/v1/" + _HUB + "/subscribers/" + seq + "/send",
							"{\"type\":\"subscribed\",\"payload\":" + t1 + "}"
					}),
					nil
			}),
		})
	}
}

func (ss *UplinkScenario) NextCall() (*Call, error) {
	for {
		if i := rand.Intn(ss.SessionAmount); i >= 0 {
			select {
			case st := <-ss._sessions[i].StepLock:
				switch st {
				case STEP1:
					// execute session call for the first time
					if ss._sessions[i]._calls[st].GenParam != nil {
						ss._sessions[i]._calls[st].Method, ss._sessions[i]._calls[st].Type, ss._sessions[i]._calls[st].URL, ss._sessions[i]._calls[st].Body = ss._sessions[i]._calls[st].GenParam()
					}

					return ss._sessions[i]._calls[st], nil
				default:
					// choose a non-initialized call randomly
					ss._sessions[i].StepLock <- REST
					q := rand.Float32() * ss._sessions[i]._totalWeight
					for j := STEP1 + 1; j < ss._sessions[i]._count; j++ {
						if q <= ss._sessions[i]._calls[j].RandomWeight {
							// add 1 to seq
							if ss._sessions[i]._calls[j].GenParam != nil {
								ss._sessions[i]._calls[j].Method, ss._sessions[i]._calls[j].Type, ss._sessions[i]._calls[j].URL, ss._sessions[i]._calls[j].Body = ss._sessions[i]._calls[j].GenParam()
							}
							return ss._sessions[i]._calls[j], nil
						}
					}
				}
			default:
				continue
			}
		}
	}

	log.Fatal("what? should never reach here")
	return nil, errors.New("all sessions are being initialized")
}

func (s *UplinkScenario) CustomizedReport() string {
	return fmt.Sprintf(" avg send: %2.5f%s", (float64(s._totalSendReceiveTime)/float64(s._totalSendCount))/1000000000, "s")
}

func init() {
	Register("uplink_session", newUplinkScenario)
}

func newUplinkScenario() (Profile, error) {
	return &UplinkScenario{
		SessionAmount: 100,
	}, nil
}
