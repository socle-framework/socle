package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	makeCmd.AddCommand(migrationCmd)
}

var migrationCmd = &cobra.Command{
	Use:   "migration <name> [fizz|sql]",
	Short: "",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		format := "fizz"
		if len(args) > 1 {
			format = strings.ToLower(args[1])
		}

		if format != "sql" && format != "fizz" {
			fmt.Println("Erreur : le format doit être 'sql' ou 'fizz'")
			os.Exit(1)
		}
		doMigration(name, format)
	},
}

func doMigration(arg3, arg4 string) error {

	checkForDB()

	//dbType := cel.DB.DataType
	if arg3 == "" {
		exitGracefully(errors.New("you must give the migration a name"))
	}

	// default to migration type of fizz
	migrationType := "fizz"
	var up, down string

	// are doing fizz or sql?
	if arg4 == "fizz" || arg4 == "" {
		upBytes, _ := templateFS.ReadFile("templates/migrations/migration_up.fizz")
		downBytes, _ := templateFS.ReadFile("templates/migrations/migration_down.fizz")

		up = string(upBytes)
		down = string(downBytes)
	} else {
		migrationType = "sql"
	}

	// create the migrations for either fizz or sql

	err := s.CreatePopMigration([]byte(up), []byte(down), arg3, migrationType)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}
