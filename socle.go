package socle

import (
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
	"github.com/socle-framework/cache"
	"github.com/socle-framework/mailer"
	"github.com/socle-framework/render"
	"github.com/socle-framework/session"
	"github.com/socle-framework/socle/pkg/auth"
	"github.com/socle-framework/socle/pkg/env"
)

const version = "0.1.2"

var redisCache *cache.RedisCache
var badgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB
var maintenanceMode bool

// New reads the .env file, creates our application config, populates the Socle type with settings
// based on .env values, and creates necessary folders and files if they don't exist
func (s *Socle) New(rootPath, entry string) error {
	s.entry = entry
	// Load .env
	err := env.Load(rootPath)
	if err != nil {
		return err
	}
	s.env = initEnvConfig()

	//load socle.yaml
	appConfig, err := LoadAppConfig(rootPath)
	if err != nil {
		return err
	}

	s.appConfig = *appConfig

	s.Debug = s.env.debug
	s.EncryptionKey = s.env.encryptionKey
	s.Version = version
	s.RootPath = rootPath

	// create loggers
	err = s.initLoggers()
	if err != nil {
		return err
	}

	// init router
	err = s.initRouter()
	if err != nil {
		return err
	}

	// connect to database
	if s.appConfig.Store.Enabled {
		err = s.initDB()
		if err != nil {
			return err
		}
	}

	// config scheduler
	err = s.initScheduler()
	if err != nil {
		return err
	}

	// cache setting
	err = s.initCache()
	if err != nil {
		return err
	}

	// init server
	err = s.initServer()
	if err != nil {
		return err
	}

	// create session
	err = s.InitSession()
	if err != nil {
		return err
	}

	//init render
	err = s.initRenderer()
	if err != nil {
		return err
	}

	// init Mailer
	err = s.initMailer()
	if err != nil {
		return err
	}
	go s.Mail.ListenForMail()

	//init auth
	err = s.initAuthentificator()
	if err != nil {
		return err
	}

	return nil
}

func (s *Socle) initLoggers() error {
	s.Log.InfoLog = log.New(os.Stdout, s.entry+" INFO\t", log.Ldate|log.Ltime)
	s.Log.ErrorLog = log.New(os.Stdout, s.entry+" ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func (s *Socle) initRouter() error {
	if !InArrayStr(s.entry, []string{"web", "api/rest"}) {
		return nil
	}
	var middlewares []string
	switch s.entry {
	case "api/rest":
		middlewares = s.appConfig.Entries.Api.Middlewares

	default:
		middlewares = s.appConfig.Entries.Web.Middlewares
	}
	s.Server.Middlewares = middlewares
	s.Routes = s.routes(middlewares).(*chi.Mux)
	return nil
}

func (s *Socle) initDB() error {
	if s.env.db.dbType != "" {
		db, err := s.OpenDB(s.env.db.dbType, s.BuildDSN())
		if err != nil {
			s.Log.ErrorLog.Println(err)
			return err
		}
		s.DB = Database{
			DBType: s.env.db.dbType,
			Pool:   db,
		}
	}

	return nil
}

func (s *Socle) initScheduler() error {
	s.Scheduler = cron.New()
	return nil
}

func (s *Socle) initCache() error {
	if s.env.cache == "redis" || s.env.sessionType == "redis" {
		redisCache = s.createClientRedisCache()
		s.Cache = redisCache
		redisPool = redisCache.Conn
	}

	if s.env.cache == "badger" {
		badgerCache = s.createClientBadgerCache()
		s.Cache = badgerCache
		badgerConn = badgerCache.Conn

		_, err := s.Scheduler.AddFunc("@daily", func() {
			_ = badgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Socle) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   s.createRedisPool(),
		Prefix: s.env.redis.prefix,
	}
	return &cacheClient
}

func (s *Socle) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: s.createBadgerConn(),
	}
	return &cacheClient
}

func (s *Socle) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				s.env.redis.host,
				redis.DialPassword(s.env.redis.password))
		},

		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}

func (s *Socle) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(s.RootPath + "/tmp/badger"))
	if err != nil {
		return nil
	}
	return db
}

func (s *Socle) initMailer() error {
	s.Mail = mailer.Mail{
		Domain:      s.env.mail.domain,
		Templates:   s.RootPath + "/mail",
		Host:        s.env.mail.smtp.host,
		Port:        s.env.mail.smtp.port,
		Username:    s.env.mail.smtp.username,
		Password:    s.env.mail.smtp.password,
		Encryption:  s.env.mail.smtp.encryptionKey,
		FromName:    s.env.mail.fromName,
		FromAddress: s.env.mail.FromAddress,
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         s.env.mail.mailerService.api,
		APIKey:      s.env.mail.mailerService.key,
		APIUrl:      s.env.mail.mailerService.url,
	}

	return nil
}

func (s *Socle) initServer() error {
	if !InArrayStr(s.entry, []string{"web", "api/rest"}) {
		return nil
	}

	s.Server = Server{
		Name:    s.env.serverName,
		Address: s.env.serverAddress,
	}
	switch s.entry {
	case "api/rest":
		s.Server.Port = s.env.restApiPort
		s.Server.Secure = s.appConfig.Entries.Api.Security.Enabled
		s.Server.Security.Strategy = s.appConfig.Entries.Api.Security.TLS.Strategy
		s.Server.Security.MutualTLS = s.appConfig.Entries.Api.Security.TLS.Mutual
		s.Server.Security.CAName = s.appConfig.Entries.Api.Security.TLS.CACertName
		s.Server.Security.ServerCertName = s.appConfig.Entries.Api.Security.TLS.ServerCertName
		s.Server.Security.ClientCertName = s.appConfig.Entries.Api.Security.TLS.ClientCertName

	default:
		s.Server.Port = s.env.webPort
		s.Server.Secure = s.appConfig.Entries.Web.Security.Enabled
		s.Server.Security.Strategy = s.appConfig.Entries.Web.Security.TLS.Strategy
		s.Server.Security.MutualTLS = s.appConfig.Entries.Web.Security.TLS.Mutual
		s.Server.Security.CAName = s.appConfig.Entries.Web.Security.TLS.CACertName
		s.Server.Security.ServerCertName = s.appConfig.Entries.Web.Security.TLS.ServerCertName
		s.Server.Security.ClientCertName = s.appConfig.Entries.Web.Security.TLS.ClientCertName
	}
	return nil

}

func (s *Socle) InitSession() error {
	// if s.entry != "web" {
	// 	return nil
	// }

	sess := session.Session{
		CookieLifetime: s.env.cookie.lifetime,
		CookiePersist:  s.env.cookie.persist,
		CookieName:     s.env.cookie.name,
		SessionType:    s.env.sessionType,
	}

	if s.env.cookie.domain != "" {
		sess.CookieDomain = s.env.cookie.domain
	}
	switch s.env.sessionType {
	case "redis":
		sess.RedisPool = redisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = s.DB.Pool
	}

	s.Session = sess.InitSession()
	return nil
}

func (s *Socle) initRenderer() error {
	if s.entry != "web" {
		return nil
	}

	switch s.appConfig.Entries.Web.Render {
	case "templ":
		rd := &render.TemplRender{}
		rd.RootPath = s.RootPath
		rd.Session = s.Session
		s.Render = rd

	// 		s.Render = &render.Render{
	// 	Renderer: s.appConfig.Entries.Web.Render,
	// 	RootPath: s.RootPath,
	// 	Session:  s.Session,
	// }
	case "jet":
		rd := &render.JetRender{}
		rd.RootPath = s.RootPath
		rd.Session = s.Session
		s.Render = rd
	default:

	}
	return nil
}

func (s *Socle) initAuthentificator() error {
	if s.entry != "api/rest" {
		return nil
	}

	s.Authenticator = auth.NewJWTAuthenticator(
		s.env.auth.token.secret,
		s.env.auth.token.iss,
		s.env.auth.token.iss,
	)

	return nil
}

type RPCServer struct{}

func (r *RPCServer) MaintenanceMode(inMaintenanceMode bool, resp *string) error {
	if inMaintenanceMode {
		maintenanceMode = true
		*resp = "Server in maintenance mode"
	} else {
		maintenanceMode = false
		*resp = "Server live!"
	}
	return nil
}

func (s *Socle) listenRPC() {
	// if nothing specified for rpc port, don't start
	if os.Getenv("RPC_PORT") != "" {
		s.Log.InfoLog.Println("Starting RPC server on port", os.Getenv("RPC_PORT"))
		err := rpc.Register(new(RPCServer))
		if err != nil {
			s.Log.ErrorLog.Println(err)
			return
		}
		listen, err := net.Listen("tcp", "127.0.0.1:"+os.Getenv("RPC_PORT"))
		if err != nil {
			s.Log.ErrorLog.Println(err)
			return
		}
		for {
			rpcConn, err := listen.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(rpcConn)
		}

	}
}
