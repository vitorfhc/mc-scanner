package cmd

type CliParams struct {
	MaxJobs      int
	AddrListFile string
	Timeout      int
	Debug        bool
}

var GlobalCliParams = &CliParams{}
