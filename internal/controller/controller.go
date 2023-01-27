package controller

import (
	"context"
	"errors"
	"sync"
	"time"
)

type controller struct {
	inputs           chan string
	outputs          chan string
	options          *ControllerOptions
	workers          []Worker
	closeInputsOnce  sync.Once
	closeOutputsOnce sync.Once
	NumInputs        int
	NumOutputs       int
	NumErrors        int
}

type ControllerOptions struct {
	RequestTimeout       int
	MaxConcurrentWorkers int
}

func New() *controller {
	defaultMaxWorkers := 10

	ctl := &controller{
		inputs:  make(chan string, 1000), // TODO: review this hardcoded 1000
		outputs: make(chan string, 1000),
		workers: make([]Worker, 0, defaultMaxWorkers),
		options: &ControllerOptions{
			RequestTimeout:       10,
			MaxConcurrentWorkers: defaultMaxWorkers,
		},
	}

	return ctl
}

func NewWithOptions(options *ControllerOptions) *controller {
	ctl := &controller{
		inputs:  make(chan string, 1000),
		outputs: make(chan string, 1000),
		workers: make([]Worker, 0, options.MaxConcurrentWorkers),
		options: options,
	}

	return ctl
}

func (ctlr *controller) Outputs() chan string {
	return ctlr.outputs
}

func (ctlr *controller) Inputs() chan string {
	return ctlr.inputs
}

func (ctlr *controller) RegisterWorker(w Worker) error {
	size := len(ctlr.workers)
	if size >= cap(ctlr.workers) {
		return errors.New("controller: RegisterWorker: can't add more workers than the limit")
	}

	ctlr.workers = append(ctlr.workers, w)

	return nil
}

func (ctl *controller) RunWorkers(ctx context.Context) error {
	// Check if we have nil workers
	for _, w := range ctl.workers {
		if w == nil {
			return errors.New("controller: RunWorkers: must register all the workers")
		}
	}

	var workersWg sync.WaitGroup
	workersCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	workerOptions := ctl.buildWorkerOptions()

	// Get all the workers running
	for _, w := range ctl.workers {
		workersWg.Add(1)
		go func(w Worker) {
			defer workersWg.Done()
			w.Run(workersCtx, workerOptions)
		}(w)
	}

	// Check for cancels
	var checkerWg sync.WaitGroup
	checkerCtx, cancelChecker := context.WithCancel(context.Background())
	checkerWg.Add(1)
	go func() {
		defer checkerWg.Done()
		for {
			select {
			case <-ctx.Done(): // the caller asked the workers to stop
				cancelWorkers()
				return
			case <-workersCtx.Done(): // someone inside asked the workers to stop
				return
			case <-checkerCtx.Done(): // time to stop the checker
				return
			default:
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}()

	workersWg.Wait()
	cancelChecker() // stop checker if all workers are stopped
	checkerWg.Wait()

	return nil
}

func (ctlr *controller) CloseInputs() {
	ctlr.closeInputsOnce.Do(func() {
		close(ctlr.inputs)
	})
}

func (ctlr *controller) CloseOutputs() {
	ctlr.closeOutputsOnce.Do(func() {
		close(ctlr.outputs)
	})
}

func (ctlr *controller) buildWorkerOptions() *WorkerOptions {
	wo := &WorkerOptions{
		RequestTimeout: ctlr.options.RequestTimeout,
		Inputs:         ctlr.inputs,
		Outputs:        ctlr.outputs,
		Controller:     ctlr,
	}

	return wo
}
