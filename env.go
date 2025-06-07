package socle

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/socle-framework/socle/pkg/env"
	"github.com/socle-framework/socle/pkg/ratelimiter"
)

type envConfig struct {
	appName       string
	mode          string
	debug         bool
	webPort       string
	apiPort       string
	rpcPort       string
	serverName    string
	secure        bool
	db            dbConfig
	auth          authConfig
	redis         redisConfig
	cache         string
	cookie        cookieConfig
	sessionType   string
	mail          mailConfig
	migrationUrl  string
	rateLimiter   ratelimiter.Config
	uploads       uploadConfig
	encryptionKey string
	storage       storageConfig
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret  string
	exp     time.Duration
	refresh time.Duration
	iss     string
}

type basicConfig struct {
	user string
	pass string
}

type dbConfig struct {
	dbType       string
	host         string
	port         string
	user         string
	pass         string
	name         string
	ssl          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

// cookieConfig holds cookie config values
type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

type redisConfig struct {
	host     string
	password string
	prefix   string
	db       int
	enabled  bool
}

type Logger struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
}

type uploadConfig struct {
	allowedMimeTypes []string
	maxUploadSize    int64
}

type mailConfig struct {
	smtp          smtpConfig
	mailerService mailerServiceConfig
	domain        string
	fromName      string
	FromAddress   string
}

type smtpConfig struct {
	host          string
	username      string
	password      string
	port          int
	encryptionKey string
}

type mailerServiceConfig struct {
	api string
	key string
	url string
}

type storageConfig struct {
	s3     s3Config
	minio  minioConfig
	sftp   sftpConfig
	webDAV webdavConfig
}

type s3Config struct {
	secret   string
	key      string
	region   string
	endpoint string
	bucket   string
}

type minioConfig struct {
	endpoint string
	key      string
	secret   string
	useSSL   bool
	region   string
	bucket   string
}

type sftpConfig struct {
	host string
	user string
	pass string
	port string
}

type webdavConfig struct {
	host string
	user string
	pass string
}

func initEnvConfig() envConfig {

	exploded := strings.Split(os.Getenv("ALLOWED_FILETYPES"), ",")
	var mimeTypes []string
	for _, m := range exploded {
		mimeTypes = append(mimeTypes, m)
	}

	var maxUploadSize int64
	if max, err := strconv.Atoi(os.Getenv("MAX_UPLOAD_SIZE")); err != nil {
		maxUploadSize = 10 << 20
	} else {
		maxUploadSize = int64(max)
	}

	return envConfig{
		appName:      env.GetString("APP_NAME", ""),
		mode:         env.GetString("MODE", "dev"),
		debug:        env.GetBool("DEBUG", true),
		webPort:      env.GetString("WEB_PORT", "8090"),
		apiPort:      env.GetString("API_PORT", "8091"),
		rpcPort:      env.GetString("RPC_PORT", "8092"),
		serverName:   env.GetString("SERVER_NAME", "localhost"),
		secure:       env.GetBool("SECURE", false),
		cache:        env.GetString("CACHE", "memory"),
		sessionType:  env.GetString("SESSION_TYPE", "cookie"),
		migrationUrl: env.GetString("MIGRATION_URL", "file://migrate/migration"),
		db: dbConfig{
			dbType:       env.GetString("DATABASE_TYPE", "postgres"),
			host:         env.GetString("DATABASE_HOST", "localhost"),
			port:         env.GetString("DATABASE_PORT", "5432"),
			user:         env.GetString("DATABASE_USER", "postgres"),
			pass:         env.GetString("DATABASE_PASS", "password"),
			name:         env.GetString("DATABASE_NAME", "dbname"),
			ssl:          env.GetString("DATABASE_SSL", "disable"),
			maxOpenConns: env.GetInt("DATABASE_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DATABASE_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DATABASE_MAX_IDLE_TIME", "15m"),
		},

		redis: redisConfig{
			host:     env.GetString("REDIS_ADDR", "localhost:6379"),
			password: env.GetString("REDIS_PW", ""),
			prefix:   env.GetString("REDIS_PREFIX", ""),
			db:       env.GetInt("REDIS_DB", 0),
			enabled:  env.GetBool("REDIS_ENABLED", false),
		},

		cookie: cookieConfig{
			name:     env.GetString("COOKIE_NAME", "app"),
			lifetime: env.GetString("COOKIE_LIFETIME", "24h"),
			persist:  env.GetString("COOKIE_PERSIST", "true"),
			secure:   env.GetString("COOKIE_SECURE", "false"),
			domain:   env.GetString("COOKIE_DOMAIN", "localhost"),
		},

		mail: mailConfig{
			FromAddress: env.GetString("FROM_EMAIL", ""),
			fromName:    env.GetString("FROM_NAME", ""),
			domain:      env.GetString("MAIL_DOMAIN", "localhost"),
			smtp: smtpConfig{
				host:     env.GetString("SMTP_HOST", ""),
				port:     env.GetInt("SMTP_PORT", 587),
				username: env.GetString("SMTP_USER", ""),
				password: env.GetString("SMTP_PASS", ""),
			},
			mailerService: mailerServiceConfig{
				api: env.GetString("MAILER_API", "sendgrid"),
				key: env.GetString("MAILER_KEY", ""),
				url: env.GetString("MAILER_URL", ""),
			},
		},

		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret:  env.GetString("AUTH_TOKEN_SECRET", ""),
				exp:     time.Hour * time.Duration(env.GetInt("AUTH_TOKEN_EXP", 72)),     // 1 day
				refresh: time.Hour * time.Duration(env.GetInt("AUTH_TOKEN_REFRESH", 72)), // 3 days
				iss:     env.GetString("AUTH_TOKEN_ISS", "app"),
			},
		},

		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATE_LIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Second * time.Duration(env.GetInt("RATE_LIMITER_TIME", 72)),
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", false),
		},
		encryptionKey: env.GetString("KEY", "default-key-should-be-32-bytes!"),
		uploads: uploadConfig{
			allowedMimeTypes: mimeTypes,
			maxUploadSize:    maxUploadSize, // 10 MB par défaut
		},

		storage: storageConfig{
			s3: s3Config{
				secret:   env.GetString("S3_SECRET", ""),
				key:      env.GetString("S3_KEY", ""),
				region:   env.GetString("S3_REGION", ""),
				endpoint: env.GetString("S3_ENDPOINT", ""),
				bucket:   env.GetString("S3_BUCKET", ""),
			},
			minio: minioConfig{
				endpoint: env.GetString("MINIO_ENDPOINT", "127.0.0.1:8008"),
				key:      env.GetString("MINIO_KEY", "root"),
				secret:   env.GetString("MINIO_SECRET", "password"),
				useSSL:   env.GetBool("MINIO_USESSL", false),
				region:   env.GetString("MINIO_REGION", "togo-lome-1"),
				bucket:   env.GetString("MINIO_BUCKET", "testbucket"),
			},
			sftp: sftpConfig{
				host: env.GetString("SFTP_HOST", ""),
				user: env.GetString("SFTP_USER", ""),
				pass: env.GetString("SFTP_PASS", ""),
				port: env.GetString("SFTP_PORT", ""),
			},
			webDAV: webdavConfig{
				host: env.GetString("WEBDAV_HOST", ""),
				user: env.GetString("WEBDAV_USER", ""),
				pass: env.GetString("WEBDAV_PASS", ""),
			},
		},
	}

}
