package mcscanner

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
)

func RunAsyncScannerController(options Options) {
	errors := 0
	success := 0

	var wg sync.WaitGroup

	go func() {
		for {
			logrus.Infof("Processed:\t%d\n", errors+success)
			logrus.Infof("Error:\t%d\n", errors)
			logrus.Infof("Success:\t%d\n", success)
			time.Sleep(2 * time.Second)
		}
	}()

	limit := make(chan bool, options.MaxJobs)

	for addr := range options.InputChan {
		limit <- true
		wg.Add(1)

		go func(scanAddr string) {
			defer func() { <-limit; wg.Done() }()

			bgCtx := context.Background()
			timeoutDuration := time.Duration(options.Timeout) * time.Second
			ctx, cancel := context.WithTimeout(bgCtx, timeoutDuration)
			defer cancel()

			res, err := ScanAddress(ctx, scanAddr)
			if err != nil {
				// logrus.Errorf("Error on address %q: %v\n", scanAddr, err)
				errors++
				return
			}
			success++

			options.ResultsChan <- res
		}(addr)
	}
	wg.Wait()
}

func ScanAddress(ctx context.Context, addr string) (*PingAndListResponse, error) {
	bytes, _, err := bot.PingAndListContext(ctx, addr)
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
