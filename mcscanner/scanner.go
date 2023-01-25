package mcscanner

import (
	"encoding/json"
	"time"

	"github.com/Tnze/go-mc/bot"
)

func RunAsyncScannerController(addressesChan chan string, resultsChan chan *PingAndListResponse) {
	maxWait := 4 // TODO improve this
	currWait := 0
	for {
		select {
		case addr := <-addressesChan:
			currWait = 0
			go func() {
				res, err := ScanAddress(addr, 10)
				if err != nil {
					return
				}
				resultsChan <- res
			}()
		default:
			if currWait >= maxWait {
				return
			}
			currWait = currWait + 1
			time.Sleep(time.Duration(500) * time.Millisecond)
			continue
		}
	}
}

func ScanAddress(addr string, timeout int) (*PingAndListResponse, error) {
	bytes, _, err := bot.PingAndListTimeout(addr, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}

	res := PingAndListResponse{}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}

	res.Address = addr

	return &res, nil
}
