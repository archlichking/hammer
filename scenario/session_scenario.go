package scenario

import (
	"log"
	"math/rand"
	"errors"
)

const (
	_             = iota
	STEP1  int = 0
	STEP2 int = 1
	STEP3  int = 2
	REST int = 100

	NEXT int = 1
	STAY int = 0
	PREV int = -1
)

type GenSession func() (w float32, gc GenCall, cb GenCallBack)

type SessionScenario struct {
	_sessions     []*Session
	SessionAmount int
	_count          int
}

func (ss *SessionScenario) InitFromFile(path string) {
	
}

func (ss *SessionScenario) InitFromCode() {
	ss._sessions = make([]*Session, ss.SessionAmount)
	for i := 0; i < ss.SessionAmount; i++ {
		ss.addSession([]GenSession{
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 0,
					GenCall(func(...string) (_m, _t, _u, _b string) {
						return "POST", "REST", "http://localhost:9988/post", "{\"fsdfsdfsdf\":\"ddddd\"}"
						// return "GET", "REST", "http://localhost:9988/get", "{}"
					}),
					GenCallBack(func(se *Session, st int, storage string) {
						se.UpdateStateAndStorage(st, storage)
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 0,
					GenCall(func(...string) (_m, _t, _u, _b string) {
						return "POST", "REST", "http://localhost:9988/post", "{\"fsdfsdfsdf\":\"ddddd\"}"
						// return "GET", "REST", "http://localhost:9988/get", "{}"
					}),
					GenCallBack(func(se *Session, st int, storage string) {
						se.UpdateStateAndStorage(st, storage)
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 55,
					GenCall(func(...string) (_m, _t, _u, _b string) {
						// return "POST", "REST", "http://localhost:9988/post", "{\"fsdfsdfsdf\":\"ddddd\"}"
						return "GET", "REST", "http://localhost:9988/get", "{}"
					}),
					nil
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 45,
					GenCall(func(...string) (_m, _t, _u, _b string) {
						// return "POST", "REST", "http://localhost:9988/post", "{\"fsdfsdfsdf\":\"ddddd\"}"
						return "GET", "REST", "http://localhost:9988/get", "{}"
					}),
					nil
			}),
		})
	}
}

func (ss *SessionScenario) NextCall() (*Call, error) {
	for {
		if i := rand.Intn(ss.SessionAmount); i >= 0 {
			select {
			case st := <- ss._sessions[i].StepLock :
				switch st{
				case STEP1, STEP2:
					// log.Println("step1 ", i)
					if ss._sessions[i]._calls[st].GenParam != nil {
						ss._sessions[i]._calls[st].Method, ss._sessions[i]._calls[st].Type, ss._sessions[i]._calls[st].URL, ss._sessions[i]._calls[st].Body = ss._sessions[i]._calls[st].GenParam()
					}
					// execute session call for the first time
					return ss._sessions[i]._calls[st], nil
				default:
					// choose a non-initialized call randomly
					ss._sessions[i].StepLock <- REST
					// log.Println("running", i)
					// p._sessions[i].UpdateState(NEXT)
					q := rand.Float32() * ss._sessions[i]._totalWeight
					for j := STEP1+1; j < ss._sessions[i]._count; j++ {
						if q <= ss._sessions[i]._calls[j].RandomWeight {
							if ss._sessions[i]._calls[j].GenParam != nil {
								// log.Println("fsdafdoieosr ", p._sessions[i]._calls[j].GenParam)
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


func (ss *SessionScenario) addSession(gens []GenSession) {
	// log.Println(len(gens), p._count, p._sessions)
	ss._sessions[ss._count] = new(Session)
	ss._sessions[ss._count].StepLock = make(chan int, 1)
	ss._sessions[ss._count].StepLock <- STEP1
	ss._sessions[ss._count]._calls = make([]*Call, len(gens))
	for i := 0; i < len(gens); i++ {
		w, gp, cb := gens[i]()
		ss._sessions[ss._count].addCall(w, gp, cb)
	}

	ss._count++
}

func init() {
	Register("session", newSessionScenario)
}

func newSessionScenario() (Profile, error) {
	return &SessionScenario{
		SessionAmount: 100,
	}, nil
}
