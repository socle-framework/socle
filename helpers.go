package socle

import (
	"fmt"
	"os"
)

// CreateDirIfNotExist creates a new directory if it does not exist
func (c *Socle) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFileIfNotExists creates a new file at path if it does not exist
func (c *Socle) CreateFileIfNotExists(path string) error {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}
	return nil
}

// BuildDSN builds the datasource name for our database, and returns it as a string
func (s *Socle) BuildDSN() string {
	var dsn string

	switch s.env.db.dbType {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			s.env.db.host,
			s.env.db.port,
			s.env.db.user,
			s.env.db.name,
			s.env.db.ssl)

		// we check to see if a database password has been supplied, since including "password=" with nothing
		// after it sometimes causes postgres to fail to allow a connection.
		if s.env.db.pass != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, s.env.db.pass)
		}

	case "mysql", "mariadb":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?collation=utf8_unicode_ci&timeout=5s&parseTime=true&tls=%s&readTimeout=5s",
			s.env.db.user,
			s.env.db.pass,
			s.env.db.host,
			s.env.db.port,
			s.env.db.name,
			s.env.db.ssl)

	default:

	}

	return dsn
}

func InArrayStr(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
