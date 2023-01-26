package mcscanner

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
)

func RunAsyncScannerController(addressesChan chan string, resultsChan chan *PingAndListResponse, maxJobs int) {
	errors := 0
	success := 0
	var wg sync.WaitGroup
	limit := make(chan bool, maxJobs)
	for addr := range addressesChan {
		wg.Add(1)
		limit <- true
		go func(scanAddr string) {
			defer func() { <-limit; wg.Done() }()
			res, err := ScanAddress(scanAddr, 10)
			if err != nil {
				logrus.Errorf("Error on address %q: %v\n", scanAddr, err)
				errors++
				return
			}
			success++
			resultsChan <- res
		}(addr)
	}
	wg.Wait()
	logrus.Warnf("Error on %d addresses\n", errors)
	logrus.Infof("Found %d servers\n", success)
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
