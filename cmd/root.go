package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vitorfhc/mc-scanner/internal/controller"
	"github.com/vitorfhc/mc-scanner/internal/worker"
)

var rootCmd = &cobra.Command{
	Use:   "mc-scanner",
	Short: "Scan multiple Minecraft servers in seconds",
	Long: `The mc-scanner scans multiple Minecraft servers async.
All you need is a list of available addresses in the format <address>[:<port>].

Built by Vitor Falcão <vitorfhc@protonmail.com>`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		if GlobalCliParams.Debug {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debug("debug mode is enabled")
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Open the addresses file
		filename := GlobalCliParams.AddrListFile
		logrus.Infof("opening file %q\n", filename)
		file, err := os.Open(filename)
		if err != nil {
			logrus.Fatalf("could not open file %q: %s\n", filename, err)
		}
		defer file.Close()

		// Setup the controller
		logrus.Info("setting up controller")
		ctlr := controller.NewWithOptions(&controller.ControllerOptions{
			RequestTimeout:       GlobalCliParams.Timeout,
			MaxConcurrentWorkers: GlobalCliParams.MaxJobs,
		})
		defer ctlr.CloseOutputs()
		defer ctlr.CloseInputs()

		logrus.Info("registering workers")
		for i := 0; i < GlobalCliParams.MaxJobs; i++ {
			threadWorker := worker.New()
			err := ctlr.RegisterWorker(threadWorker)
			if err != nil {
				log.Fatal("error while trying to register workers")
			}
		}

		// Start sending to the input channel
		linesCount, err := lineCounter(file)
		if err != nil {
			logrus.Fatal(err)
		}
		file.Seek(0, 0)
		ctlr.NumInputs = linesCount
		var wg sync.WaitGroup
		readCtx, cancelRead := context.WithCancel(context.Background())
		defer cancelRead()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer ctlr.CloseInputs()
			readFileToChan(readCtx, file, ctlr.Inputs())
		}()

		// Run workers
		logrus.Info("starting workers")
		workersCtx, cancelWorkers := context.WithCancel(context.Background())
		defer cancelWorkers()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer ctlr.CloseOutputs()
			defer logrus.Info("all workers finished")
			ctlr.RunWorkers(workersCtx)
		}()

		// Print output
		printCtx, cancelPrint := context.WithCancel(context.Background())
		defer cancelPrint()
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer logrus.Info("finished printing outputs")
			for {
				select {
				case <-printCtx.Done():
					return
				case output, ok := <-ctlr.Outputs():
					if !ok {
						return
					}
					fmt.Println(output)
				case <-workersCtx.Done():
					return
				default:
					time.Sleep(1 * time.Second)
					continue
				}
			}
		}()

		// Run percentager
		wg.Add(1)
		pctgCtx, cancelPctg := context.WithCancel(context.Background())
		defer cancelPctg()
		go func() {
			defer wg.Done()
			defer logrus.Info("stopped printing status")
			for {
				select {
				case <-pctgCtx.Done():
					return
				case _, ok := <-ctlr.Outputs():
					if !ok {
						return
					}
				default:
					time.Sleep(2 * time.Second)
					p := float64(ctlr.NumOutputs) / float64(ctlr.NumInputs) * 100
					pe := float64(ctlr.NumErrors) / float64(ctlr.NumInputs) * 100
					logrus.Infof("done: %d / %d = %.1f\n", ctlr.NumOutputs, ctlr.NumInputs, p)
					logrus.Infof("errors %d / %d = %.1f\n", ctlr.NumErrors, ctlr.NumInputs, pe)
				}
			}
		}()

		// Run checker for cmd context
		var checkerWg sync.WaitGroup
		checkerCtx, cancelChecker := context.WithCancel(context.Background())
		checkerWg.Add(1)
		go func() {
			defer checkerWg.Done()
			for {
				select {
				case <-cmd.Context().Done(): // SIGINT or SIGTERM
					logrus.Info("canceling all jobs")
					cancelRead()
					cancelWorkers()
					cancelPrint()
					cancelPctg()
					return
				case <-checkerCtx.Done(): // Time to leave
					return
				default:
					time.Sleep(500 * time.Millisecond)
					continue
				}
			}
		}()

		wg.Wait()
		logrus.Info("all threads stopped")
		cancelChecker()
		checkerWg.Wait()
		logrus.Info("finished")
	},
}

func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&GlobalCliParams.AddrListFile, "file", "f", "addresses.txt", "Name of the file containing a list of addresses in <address>:<port> format")
	rootCmd.Flags().IntVarP(&GlobalCliParams.MaxJobs, "max-jobs", "j", 10, "Hard limit of connections to try at the same time")
	rootCmd.Flags().IntVarP(&GlobalCliParams.Timeout, "timeout", "t", 10, "Time limit to wait for the server response in seconds")
	rootCmd.Flags().BoolVarP(&GlobalCliParams.Debug, "debug", "d", false, "Enables debug logging")
}
