package socle

import (
	"database/sql"
	"fmt"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
	"github.com/socle-framework/cache"
	"github.com/socle-framework/filesystems"
	"github.com/socle-framework/mailer"
	"github.com/socle-framework/render"
	"github.com/socle-framework/socle/pkg/auth"
	"github.com/socle-framework/socle/pkg/ratelimiter"
)

// Socle is the overall type for the Socle package. Members that are exported in this type
// are available to any application that uses it.
type Socle struct {
	appConfig     appConfig
	env           envConfig
	entry         string
	AppName       string
	Version       string
	Debug         bool
	RootPath      string
	Log           Logger
	Routes        *chi.Mux
	Render        render.Render
	Session       *scs.SessionManager
	EncryptionKey string
	Cache         cache.Cache
	DB            Database
	Authenticator auth.Authenticator
	Server        Server
	Scheduler     *cron.Cron
	Mail          mailer.Mail
	FileSystem    filesystems.FS
	RateLimiter   *ratelimiter.Limiter
}

type Database struct {
	DBType string
	Pool   *sql.DB
}

type Server struct {
	Name        string
	Address     string
	Port        string
	Secure      bool
	Security    ServerSecurity
	Middlewares []string
}

type ServerSecurity struct {
	Strategy       string
	MutualTLS      bool
	CAName         string
	ServerCertName string
	ClientCertName string
}

func (s Server) getURL() string {
	return fmt.Sprintf("%s:%s", s.Name, s.Port)
}
