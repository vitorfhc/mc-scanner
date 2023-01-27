package worker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/vitorfhc/mc-scanner/internal/api"
	"github.com/vitorfhc/mc-scanner/internal/controller"
)

type threadWorker struct{}

func New() *threadWorker {
	return &threadWorker{}
}

func (tw *threadWorker) Run(ctx context.Context, wo *controller.WorkerOptions) {
	for {
		select {
		case <-ctx.Done():
			return
		case input := <-wo.Inputs:
			output, err := scan(input, wo.RequestTimeout)
			if err == nil {
				wo.Outputs <- output
			}
		default:
			continue
		}
	}
}

func scan(addr string, timeout int) (*api.PingAndListResponse, error) {
	to := time.Duration(timeout) * time.Second
	bytes, _, err := bot.PingAndListTimeout(addr, to)
	if err != nil {
		return nil, err
	}

	res := api.PingAndListResponse{}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return nil, err
	}
	res.Address = addr

	return &res, nil
}
