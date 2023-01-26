package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vitorfhc/mc-scanner/mclog"
	"github.com/vitorfhc/mc-scanner/mcscanner"
)

var rootCmd = &cobra.Command{
	Use:   "mc-scanner",
	Short: "Scan multiple Minecraft servers in seconds",
	Long: `The mc-scanner scans multiple Minecraft servers async.
All you need is a list of available addresses in the format <address>:<port>.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger, err := mclog.New()
		if err != nil {
			logrus.Fatalf("Could not setup logger: %s\n", err)
		}

		// Enable debug mode
		if GlobalCliParams.Debug {
			logger.Debug("Debug mode is activated")
			logger.SetLevel(logrus.DebugLevel)
		}

		// Open the file with addresses
		filename := GlobalCliParams.AddrListFile
		logger.Debugf("Opening file %q\n", filename)
		file, err := os.Open(filename)
		if err != nil {
			logger.Fatalf("Could not open file %q: %s\n", filename, err)
		}
		defer file.Close()

		var wg sync.WaitGroup

		// Create the input and output channel
		addrsChan := make(chan string)
		scanner := bufio.NewScanner(file)
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Debug("Starting to populate the input channel")
			for scanner.Scan() {
				addrsChan <- scanner.Text()
			}
		}()
		resultsChan := make(chan *mcscanner.PingAndListResponse)

		// Set all options for the scanner
		options := mcscanner.Options{
			Timeout:     GlobalCliParams.Timeout,
			MaxJobs:     GlobalCliParams.MaxJobs,
			InputChan:   addrsChan,
			ResultsChan: resultsChan,
		}

		// Run the scans async
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Debug("Starting scanner")
			mcscanner.RunScanJobs(options)
			close(resultsChan)
		}()

		// Get all the incoming results
		for res := range resultsChan {
			fmt.Printf("%d/%d @ %q @ %q\n", res.Players.Online, res.Players.Max, res.Description.Text, res.Address)
		}

		// Wait scans to finish
		wg.Wait()
	},
}

func Execute() {
	err := rootCmd.Execute()
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
