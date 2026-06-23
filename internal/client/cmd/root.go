// Package cmd implements the GophKeeper CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/Irongoshan-ux/gophkeeper/pkg/version"
	"github.com/spf13/cobra"
)

var (
	serverAddr string
	insecure   bool
	masterPass string
)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper password manager CLI",
}

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "localhost:9090", "gRPC server address")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", true, "skip TLS verification (dev)")
	rootCmd.PersistentFlags().StringVar(&masterPass, "master", "", "master password for encryption")

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(registerCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(newSecretCmd())
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build version and date",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.Info())
		},
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
