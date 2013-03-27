package scenario

import (
	// "strconv"
	// "time"
	"math/rand"
	"log"
)

type AppServerScenario struct {
	SessionScenario

	_mustDone int
}

func (ass *AppServerScenario) InitFromCode() {
	_HOST := "http://23.20.148.107" // HC qa1
	_GAME := "/hc"
	// _UDID := strconv.FormatInt(time.Now().UnixNano(), 10)

	genR := []GenRequest{
		GenRequest(func(ps ...string) (_m, _t, _u, _b string) {
			// ps[0] : udid
			// ps[1] : sequence_number
			// ps[2] : playerid
			return "POST", 
				   "REST", 
				   _HOST + _GAME + "/index.php/json_gateway?svc=BatchController.authenticate_iphone", 
				   `[{"app_uuid":"` + ps[0] + `","udid":"` + ps[0] + `","mac_address":"macaddr6"},{"seconds_from_gmt":-28800,"game_name":"HCGame","client_version":"1.0","session_id":"3115749","ios_version":"iOS 5.0.1","data_connection_type":"WiFi","client_build":"10","transaction_time":"1362176918","device_type":"iPod Touch 4G","client_static_table_data":{"active":null,"using":null},"game_data_version":null},[{"_explicitType":"Command","method":"load","service":"start.game","sequence_num":` + ps[1] + `}]]`
		}),
		GenRequest(func(ps ...string) (_m, _t, _u, _b string) {
			// ps[0] : udid
			// ps[1] : sequence_number
			// ps[2] : playerid
			return "POST", 
				   "REST", 
				   _HOST + _GAME + "/index.php/json_gateway?svc=BatchController.call", 
				   `[{"_explicitType":"Session","iphone_udid":"` + ps[0] + `","start_sequence_num":"` + ps[1] + `","client_build":"10","client_version":"1.0","transaction_time":"1362768794","api_version":"1","player_id":null,"end_sequence_num":"` + ps[1] + `","game_name":"HCGame","req_id":"1","session_id":"3777470"},[{"_explicitType":"Command","params":[],"method":"finish_tutorial","service":"profile.profile","sequence_num":` + ps[1] + `}]]`
		}),
	}

	for i:=0;i<30;i++{
		genR = append(genR, GenRequest(func(ps ...string) (_m, _t, _u, _b string) {
			// ps[0] : udid
			// ps[1] : sequence_number
			// ps[2] : playerid
			return "POST", 
				   "REST", 
				   _HOST + _GAME + "/index.php/json_gateway?svc=BatchController.call", 
				   `[{"_explicitType":"Session","iphone_udid":"` + ps[0] + `","start_sequence_num":"` + ps[1] + `","client_build":"10","client_version":"1.0","transaction_time":"1360797513","api_version":"1","player_id":"` + ps[2] + `","end_sequence_num":"` + ps[1] + `","game_name":"HCGame","req_id":"1","session_id":"3777470"},[{"_explicitType":"Command","params":[],"method":"sync","service":"players.players","sequence_num":` + ps[1] + `}]]`
			}))
	}

	ass.addCallGroup(100, genR)
}

func (ass *AppServerScenario) NextCalls() (*CallGroup, int) {
	r := rand.Float32() * ass._totalWeight
	for i := 0; i < ass._count; i++ {
		if r <= ass._callGroups[i].RandomWeight {
			ass._callGroups[i].BufferedChn = make(chan string, 1)
			// for j, c := range ass._callGroups[i].Calls {
			// 	if c.GenFunc != nil {
			// 		// generate _UDID
			// 		_UDID := strconv.FormatInt(time.Now().UnixNano(), 10)
			// 		seq_num := strconv.Itoa(j)
			// 		c.Method, c.Type, c.URL, c.Body = c.GenFunc([]string{_UDID, seq_num}...)
			// 	}
			// }

			return &ass._callGroups[i], ass._mustDone
		}
	}

	log.Fatal("what? should never reach here")
	return &ass._callGroups[0], ass._mustDone
}

func init() {
	Register("appserversession", newAppServerScenario)
}

func newAppServerScenario() (Profile, error) {
	return &AppServerScenario{
		_mustDone: 2,
	}, nil
}