package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the Socle CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Socle CLI v0.1.0")
	},
}
