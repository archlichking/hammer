package scenario

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

type HCScenario struct {
	SessionScenario
	SessionAmount int
}

func (ss *HCScenario) InitFromCode() {
	ss._sessions = make([]*Session, ss.SessionAmount)

	_HOST := "http://23.20.148.107" // HC qa1
	_GAME := "/hc"

	for i := 0; i < ss.SessionAmount; i++ {
		/*
			ps[0]: sequence number
			ps[1]: player id
		*/
		_UDID := strconv.FormatInt(time.Now().UnixNano(), 10)
		ss.addSession([]GenSession{
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 0,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						return "POST",
							"REST",
							_HOST + _GAME + "/index.php/json_gateway?svc=BatchController.authenticate_iphone",
							`[{"app_uuid":"` + _UDID + `","udid":"` + _UDID + `","mac_address":"macaddr6"},` +
								`{"seconds_from_gmt":-28800,"game_name":"HCGame","client_version":"1.0","session_id":"3115749","ios_version":"iOS 5.0.1",` +
								`"data_connection_type":"WiFi","client_build":"10","transaction_time":"1362176918","device_type":"iPod Touch 4G",` +
								`"client_static_table_data":{"active":null,"using":null},"game_data_version":null},` +
								`[{"_explicitType":"Command","method":"load","service":"start.game","sequence_num":0}]]`
					}),
					GenCallBack(func(se *Session, st int, storage []byte) {
						se.InternalLock.Lock()
						defer se.InternalLock.Unlock()
						se.State += st
						se.StepLock <- se.State
						seq, _ := strconv.ParseInt(se.Storage["seq"], 10, 64)
						atomic.AddInt64(&seq, 1)
						se.Storage["seq"] = strconv.FormatInt(seq, 10)
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 0,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						return "POST",
							"REST",
							_HOST + _GAME + "/index.php/json_gateway?svc=BatchController.call",
							`[{"_explicitType":"Session","iphone_udid":"` +
								_UDID + `","start_sequence_num":"` + ps[0] +
								`","client_build":"10","client_version":"1.0","transaction_time":"1362768794","api_version":"1","player_id":null,"end_sequence_num":"` + ps[0] +
								`","game_name":"HCGame","req_id":"1","session_id":"3777470"},` +
								`[{"_explicitType":"Command","params":[],"method":"finish_tutorial","service":"profile.profile","sequence_num":` + ps[0] + `}]]`
					}),
					GenCallBack(func(se *Session, st int, storage []byte) {
						se.InternalLock.Lock()
						defer se.InternalLock.Unlock()
						// extract player_id from http body
						u := map[string]interface{}{}
						e := json.Unmarshal(storage, &u)
						if e != nil {
							panic(e)
						}
						player_id := u["metadata"].(map[string]interface{})["player"].(map[string]interface{})["player_id"]

						se.Storage["player_id"] = player_id.(string)
						// add 1 to seq number
						seq, _ := strconv.ParseInt(se.Storage["seq"], 10, 64)
						atomic.AddInt64(&seq, 1)
						se.Storage["seq"] = strconv.FormatInt(seq, 10)
						se.State += st
						se.StepLock <- se.State
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 100,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						return "POST",
							"REST",
							_HOST + _GAME + "/index.php/json_gateway?svc=BatchController.call",
							`[{"_explicitType":"Session","iphone_udid":"` +
								_UDID + `","start_sequence_num":"` + ps[0] +
								`","client_build":"10","client_version":"1.0","transaction_time":"1360797513","api_version":"1","player_id":"` +
								ps[1] + `","end_sequence_num":"` + ps[0] +
								`","game_name":"HCGame","req_id":"1","session_id":"3777470"},` +
								`[{"_explicitType":"Command","params":[],"method":"sync","service":"players.players","sequence_num":` + ps[0] + `}]]`
					}),
					nil
			}),
		})
	}
}

func (ss *HCScenario) NextCall() (*Call, error) {
	for {
		if i := rand.Intn(ss.SessionAmount); i >= 0 {
			select {
			case st := <-ss._sessions[i].StepLock:
				switch st {
				case STEP1, STEP2:
					// execute session call for the first time
					if ss._sessions[i]._calls[st].GenParam != nil {
						ss._sessions[i]._calls[st].Method, ss._sessions[i]._calls[st].Type, ss._sessions[i]._calls[st].URL, ss._sessions[i]._calls[st].Body = ss._sessions[i]._calls[st].GenParam(ss._sessions[i].Storage["seq"])
					}

					return ss._sessions[i]._calls[st], nil
				default:
					// choose a non-initialized call randomly
					ss._sessions[i].StepLock <- REST
					q := rand.Float32() * ss._sessions[i]._totalWeight
					for j := STEP2 + 1; j < ss._sessions[i]._count; j++ {
						if q <= ss._sessions[i]._calls[j].RandomWeight {
							// add 1 to seq
							ss._sessions[i].InternalLock.Lock()
							seq, _ := strconv.ParseInt(ss._sessions[i].Storage["seq"], 10, 64)
							atomic.AddInt64(&seq, 1)
							ss._sessions[i].Storage["seq"] = strconv.FormatInt(seq, 10)
							if ss._sessions[i]._calls[j].GenParam != nil {
								ss._sessions[i]._calls[j].Method, ss._sessions[i]._calls[j].Type, ss._sessions[i]._calls[j].URL, ss._sessions[i]._calls[j].Body = ss._sessions[i]._calls[j].GenParam(ss._sessions[i].Storage["seq"], ss._sessions[i].Storage["player_id"])
							}
							ss._sessions[i].InternalLock.Unlock()
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

func (s *HCScenario) CustomizedReport() string {
	return ""
}

func init() {
	Register("hc_session", newHCScenario)
}

func newHCScenario() (Profile, error) {
	return &HCScenario{
		SessionAmount: 5,
	}, nil
}
