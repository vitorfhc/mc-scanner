package controller

import (
	"context"

	"github.com/vitorfhc/mc-scanner/internal/api"
)

type WorkerOptions struct {
	RequestTimeout int
	Inputs         chan string
	Outputs        chan *api.PingAndListResponse
}

type Worker interface {
	Run(context.Context, *WorkerOptions)
}
