package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vitorfhc/mc-scanner/cmd"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-sigChan
		cancel()
		logrus.Info("trying to stop gracefully - waiting 10 seconds")
		time.Sleep(10 * time.Second)
		logrus.Error("unable to stop gracefully")
		os.Exit(1)
	}()

	cmd.Execute(ctx)
}
