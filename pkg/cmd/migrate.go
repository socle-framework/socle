package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate [up|down|reset] [all]",
	Short: "",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var arg1, arg2 string
		arg1 = "up"
		if len(args) > 0 {
			arg1 = args[0]
		}

		if len(args) > 1 {
			arg2 = args[1]
		}
		doMigrate(arg1, arg2)
	},
}

func doMigrate(arg1, arg2 string) error {
	//dsn := getDSN()
	checkForDB()

	tx, err := s.PopConnect()
	if err != nil {
		exitGracefully(err)
	}
	defer tx.Close()

	// run the migration command
	switch arg1 {
	case "up":
		err := s.RunPopMigrations(tx)
		if err != nil {
			return err
		}

	case "down":
		if arg2 == "all" {
			err := s.PopMigrateDown(tx, -1)
			if err != nil {
				return err
			}
		} else {
			err := s.PopMigrateDown(tx, 1)
			if err != nil {
				return err
			}
		}

	case "reset":
		err := s.PopMigrateReset(tx)
		if err != nil {
			return err
		}
	default:
		showHelp()
	}

	return nil
}
