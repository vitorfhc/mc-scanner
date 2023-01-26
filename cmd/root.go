package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vitorfhc/mc-scanner/mcscanner"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mc-scanner",
	Short: "Scan multiple Minecraft servers in seconds",
	Long: `The mc-scanner scans multiple Minecraft servers async.
All you need is a list of available addresses in the format <address>:<port>.`,
	Run: func(cmd *cobra.Command, args []string) {
		if GlobalCliParams.Debug {
			logrus.Debug("Debug mode is activated")
			logrus.SetLevel(logrus.DebugLevel)
		}

		filename := GlobalCliParams.AddrListFile
		file, err := os.Open(filename)
		if err != nil {
			logrus.Fatalf("Could not open file %q: %v\n", filename, err)
		}
		defer file.Close()

		var addresses []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			addresses = append(addresses, scanner.Text())
		}

		// timeout := GlobalCliParams.Timeout
		maxJobs := GlobalCliParams.MaxJobs
		addrsChan := make(chan string, maxJobs)
		resultsChan := make(chan *mcscanner.PingAndListResponse)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, addr := range addresses {
				addrsChan <- addr
			}
			close(addrsChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			mcscanner.RunAsyncScannerController(addrsChan, resultsChan, maxJobs)
			close(resultsChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range resultsChan {
				fmt.Printf("%d/%d @ %q @ %q\n", res.Players.Online, res.Players.Max, res.Description.Text, res.Address)
			}
		}()

		wg.Wait()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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
