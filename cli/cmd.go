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

func Init() error {
	rootCmd.PersistentFlags().StringVar(&endpoint, "endpoint", "localhost:10000", "upchain server endpoint")
	rootCmd.PersistentFlags().BoolVar(&secureConn, "secure", false, "connect with TLS")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(appendCmd)
	rootCmd.AddCommand(digestCmd)
	rootCmd.AddCommand(proofCmd)
	rootCmd.AddCommand(registerCmd)

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
