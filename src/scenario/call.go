package scenario

import (
	"errors"
	// "log"
	"eReceipts_server_service"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Item struct {
	Upc   string
	Name  string
	Price int
}

type Call struct {
	Weight   float32
	Header   map[string]string
	URL      string
	Method   string
	BodyType string
	Body     string
	/** ====== only for ereceipt */
	Host    string
	Port    string
	Store   int
	Items   []*Item
	Range   int
	Receipt *eReceipts_server_service.ISPReceipt
	/** ====== only for ereceipt */
}

func (c *Call) GenReceipt() *eReceipts_server_service.ISPReceipt {
	c.Receipt = new(eReceipts_server_service.ISPReceipt)
	// need to update this
	ruid := "1312313adsf3qrasdfv312asf"
	isMsco := true
	templateId := "AS01"
	businessType := "WMSC"
	now := uint32(time.Now().Unix())
	sms := false
	registerId := strconv.Itoa(rand.Intn(50))
	c.Receipt.IsMsco = &isMsco
	c.Receipt.Ruid = &ruid
	c.Receipt.TemplateId = &templateId
	c.Receipt.BusinessType = &businessType
	c.Receipt.CustomerWantsSms = &sms
	c.Receipt.IspTimestamp = &now
	c.Receipt.FepTimestamp = &now
	c.Receipt.RegisterId = &registerId

	c.addReceiptLine("          ( 770 ) 502 - 0677          ", 32768, false, false, 0)

	return c.Receipt
}

func (c *Call) addReceiptLine(text string, mask int32, isDoubleSize bool, isBarCode bool, lineAdvance int32) {
	rl := new(eReceipts_server_service.ReceiptLine)

	rl.Text = &text
	rl.Mask = &mask
	rl.LineAdvance = &lineAdvance
	rl.IsBarcode = &isBarCode
	rl.IsDoubleSize = &isDoubleSize

	c.Receipt.ReceiptLines = append(c.Receipt.ReceiptLines, rl)
}

type Session struct {
	Calls []*Call
	Lock  chan int
}

func (s *Session) LockNext(current int) {
	if current >= len(s.Calls)-1 {
		s.Lock <- 0
	} else {
		s.Lock <- current + 1
	}

}

type Group struct {
	Weight   float32
	Calls    []*Call
	Sessions []*Session
}

type Scenario struct {
	Groups []*Group
	Weight float32
	Type   string
}

func (s *Scenario) NextAvailable(r float32, rg *rand.Rand) (*Call, *Session, int, error) {
	switch strings.ToUpper(s.Type) {
	case "CALL":
		for j := 0; j < len(s.Groups); j++ {
			if r <= s.Groups[j].Weight {
				return s.Groups[j].Calls[0], nil, -1, nil
			}
		}
		break
	case "SESSION":
		// weight wont work in session Scenario
		for j := 0; j < len(s.Groups); j++ {
			if r <= s.Groups[j].Weight {
				for k := 0; k < 100; k++ {
					select {
					case cur, ok := <-s.Groups[j].Sessions[k].Lock:
						if ok {
							return s.Groups[j].Sessions[k].Calls[cur], s.Groups[j].Sessions[k], cur, nil
						} else {
							continue
						}
					default:
						continue
					}
				}
				return nil, nil, -1, errors.New("No session available, skip one tick")
			}
		}

		break

	}
	return nil, nil, -1, errors.New("Scenario.Type should be [call|session]")
}
