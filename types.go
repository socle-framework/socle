package socle

import (
	"database/sql"
	"log"
)

// initPaths is used when initializing the application. It holds the root
// path for the application, and a slice of strings with the names of
// folders that the application expects to find.
type initPaths struct {
	rootPath            string
	rootFolderNames     []string
	cmdFolderNames      []string
	internalFolderNames []string
	varFolderNames      []string
}

// cookieConfig holds cookie config values
type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

type databaseConfig struct {
	dsn      string
	database string
}

type Database struct {
	DataType string
	Pool     *sql.DB
}

type redisConfig struct {
	host     string
	password string
	prefix   string
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type Logger struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
	uploads     uploadConfig
}

type uploadConfig struct {
	allowedMimeTypes []string
	maxUploadSize    int64
}
