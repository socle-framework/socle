package socle

import (
	"fmt"
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
func (s *Socle) New(rootPath, cmd string) error {
	s.cmd = cmd
	// Load .env
	err := env.Load(rootPath)
	if err != nil {
		return err
	}
	s.env = initEnvConfig()
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
	err = s.initDB()
	if err != nil {
		return err
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

	return nil
}

func (s *Socle) initLoggers() error {
	s.Log.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	s.Log.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

func (s *Socle) initRouter() error {
	if !InArrayStr(s.cmd, []string{"web", "api"}) {
		return nil
	}
	s.Routes = s.routes().(*chi.Mux)
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

	if !InArrayStr(s.cmd, []string{"web", "api", "rpc"}) {
		return nil
	}

	s.Server = Server{
		ServerName: s.env.serverName,
		Secure:     s.env.secure,
	}
	switch s.cmd {
	case "api":
		s.Server.Port = s.env.apiPort

	case "rpc":
		s.Server.Port = s.env.rpcPort
	default:
		s.Server.Port = s.env.webPort
	}
	s.Server.URL = fmt.Sprintf("%s:%s", s.env.serverName, s.Server.Port)
	return nil

}

func (s *Socle) InitSession() error {
	if s.cmd != "web" {
		return nil
	}

	sess := session.Session{
		CookieLifetime: s.env.cookie.lifetime,
		CookiePersist:  s.env.cookie.persist,
		CookieName:     s.env.cookie.name,
		SessionType:    s.env.sessionType,
		CookieDomain:   s.env.cookie.domain,
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
	if s.cmd != "web" {
		return nil
	}

	s.Render = &render.Render{
		RootPath: s.RootPath,
		Session:  s.Session,
	}

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
