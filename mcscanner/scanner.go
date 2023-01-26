package mcscanner

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Tnze/go-mc/bot"
)

func RunScanJobs(options Options) {
	var wg sync.WaitGroup
	limit := make(chan bool, options.MaxJobs)

	for addr := range options.InputChan {
		limit <- true
		wg.Add(1)

		go func(scanAddr string) {
			defer func() { <-limit; wg.Done() }()

			bgCtx := context.Background()
			timeoutDuration := time.Duration(options.Timeout) * time.Second
			scanCtx, cancel := context.WithTimeout(bgCtx, timeoutDuration)
			defer cancel()

			res, err := scanAddress(scanCtx, scanAddr)
			if err != nil {
				return
			}
			options.ResultsChan <- res
		}(addr)
	}
	wg.Wait()
}

func scanAddress(ctx context.Context, addr string) (*PingAndListResponse, error) {
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
