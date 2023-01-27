package controller

import (
	"context"
)

type WorkerOptions struct {
	RequestTimeout int
	Inputs         chan string
	Outputs        chan string
	Controller     *controller
}

type Worker interface {
	Run(context.Context, *WorkerOptions)
}
