package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
)

func getDSN() string {
	dbType := s.DB.DataType

	if dbType == "pgx" {
		dbType = "postgres"
	}

	if dbType == "postgres" {
		var dsn string
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASS"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		} else {
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s",
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_SSL_MODE"))
		}
		return dsn
	}
	return "mysql://" + s.BuildDSN()
}

func checkForDB() {
	dbType := s.DB.DataType

	if dbType == "" {
		exitGracefully(errors.New("no database connection provided in .env"))
	}

	if !fileExists(s.RootPath + "/config/database.yml") {
		exitGracefully(errors.New("config/database.yml does not exist"))
	}
}

func showHelp() {
	color.Yellow(`Available commands:

	help                           - show the help commands
	down                           - put the server into maintenance mode
	up                             - take the server out of maintenance mode
	version                        - print application version
	migrate                        - runs all up migrations that have not been run previously
	migrate down                   - reverses the most recent migration
	migrate reset                  - runs all down migrations in reverse order, and then all up migrations
	make migration <name> <format> - creates two new up and down migrations in the migrations folder; format=sql/fizz (default fizz)
	make auth                      - creates and runs migrations for authentication tables, and creates models and middleware
	make handler <name>            - creates a stub handler in the handlers directory
	make model <name>              - creates a new model in the data directory
	make session                   - creates a table in the database as a session store
	make mail <name>               - creates two starter mail templates in the mail directory
	
	`)
}
