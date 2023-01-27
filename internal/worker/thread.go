package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
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
		case input, ok := <-wo.Inputs:
			if !ok {
				return
			}
			output, err := scan(input, wo.RequestTimeout)
			wo.Controller.NumOutputs++
			if err == nil {
				wo.Outputs <- output
			} else {
				wo.Controller.NumErrors++
			}
		default:
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func scan(addr string, timeout int) (string, error) {
	debugMsg := fmt.Sprintf("scanning address %q:", addr)
	timeoutDuration := time.Duration(timeout) * time.Second
	bytes, delay, err := bot.PingAndListTimeout(addr, timeoutDuration)
	if err != nil {
		debugMsg = fmt.Sprintf("%s error: %s\n", debugMsg, err)
		logrus.Debug(debugMsg)
		return "", err
	}

	debugMsg = fmt.Sprintf("%s success: delay %s\n", debugMsg, delay)
	logrus.Debug(debugMsg)

	bytesStr := string(bytes)
	finalOutput := fmt.Sprintf("{\"address\":%q, \"response\":%s}", addr, bytesStr)
	return finalOutput, nil
}
