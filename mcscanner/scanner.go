package mcscanner

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Tnze/go-mc/bot"
)

func RunAsyncScannerController(addressesChan chan string, resultsChan chan *PingAndListResponse) {
	errors := 0
	var wg sync.WaitGroup
	for addr := range addressesChan {
		wg.Add(1)
		go func(scanAddr string) {
			defer wg.Done()
			res, err := ScanAddress(scanAddr, 10)
			if err != nil {
				errors++
				return
			}
			resultsChan <- res
		}(addr)
	}
	wg.Wait()
	fmt.Println(errors)
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
