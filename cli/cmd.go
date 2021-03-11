package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	endpoint   string
	secureConn bool
)

var rootCmd = &cobra.Command{
	Use:   "upcli",
	Short: "Upcli is a simple command-line tool for upchain interaction",
}

// Init initiates commands
func Init() error {
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "localhost:10000", "upchain server endpoint")
	rootCmd.PersistentFlags().BoolVar(&secureConn, "secure", false, "connect with TLS")

	rootCmd.AddCommand(appendCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(digestCmd)
	rootCmd.AddCommand(proofCmd)
	rootCmd.AddCommand(registerCmd)

	return nil
}

// Execute executes command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
