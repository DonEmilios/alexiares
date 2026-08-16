// Command alexiares is the Alexiares infrastructure intelligence CLI.
package main

import "os"

func main() {
	// rootCmd.Execute already prints the error itself (SilenceErrors
	// is false, the default): printing it again here would duplicate
	// every error message. Only the exit code is our job.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
