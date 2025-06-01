package cmd

import (
	"github.com/danielkeho/crypto/pkg/random"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	makeCmd.AddCommand(keyCmd)
}

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "",
	Run: func(cmd *cobra.Command, args []string) {
		rnd := random.RandomString(32)
		color.Yellow("32 character encryption key: %s", rnd)
	},
}
