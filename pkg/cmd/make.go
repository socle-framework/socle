package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(makeCmd)
}

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "make a new component (handler, usecase, model)",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: call internal/scaffolder logic
	},
}
