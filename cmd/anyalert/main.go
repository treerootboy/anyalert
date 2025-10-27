package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anyalert",
	Short: "AnyAlert - A notification framework for multiple channels",
	Long: `AnyAlert is a notification framework that integrates multiple channels
like Slack, SMS, Phone, and Youdu IM, providing HTTP and gRPC service interfaces.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
