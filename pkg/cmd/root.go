package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/socle-framework/socle"
	"github.com/spf13/cobra"
)

var s socle.Socle

var rootCmd = &cobra.Command{
	Use:   "socle",
	Short: "Socle is a Go meta framework for building applications.",
	Long: `Socle CLI helps you scaffold and manage projects built on the Socle framework.

Available Commands: 
	add         				   - add web, api, rpc and grpc entries
	version     				   - Print the CLI version
	help                           - show the help commands
	down                           - put the server into maintenance mode
	up                             - take the server out of maintenance mode
	version                        - print application version
	migrate                        - runs all up migrations that have not been run previously
	migrate down                   - reverses the most recent migration
	migrate reset                  - runs all down migrations in reverse order, and then all up migrations
	make        				   - Generate handlers, models, usecases and more
	make migration <name> <format> - creates two new up and down migrations in the migrations folder; format=sql/fizz (default fizz)
	make auth                      - creates and runs migrations for authentication tables, and creates models and middleware
	make handler <name>            - creates a stub handler in the handlers directory
	make model <name>              - creates a new model in the data directory
	make session                   - creates a table in the database as a session store
	make mail <name>               - creates two starter mail templates in the mail directory

Examples:
  socle new myapp
  socle add api
  socle make handler User
  socle version
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	color.Cyan("🚀 Ready to build with Socle!")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	err := godotenv.Load()
	if err != nil {
		exitGracefully(err)
	}

	path, err := os.Getwd()
	if err != nil {
		exitGracefully(err)
	}

	s.RootPath = path
	s.DB.DBType = os.Getenv("DATABASE_TYPE")
}

func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}

	if err != nil {
		color.Red("Error: %v\n", err)
	}

	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Green("Finished!")
	}

	os.Exit(0)
}
