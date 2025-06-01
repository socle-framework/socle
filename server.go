package socle

import (
	"log"
	"net/http"
	"os"
	"time"
)

func (c *Socle) ListenAndServe() error {
	log.Printf("WEB is listening on port  %v", os.Getenv("WEB_URL"))
	srv := &http.Server{
		Addr:         os.Getenv("WEB_URL"),
		ErrorLog:     c.Log.ErrorLog,
		Handler:      c.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	if c.DB.Pool != nil {
		defer c.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	go c.listenRPC()
	c.Log.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))
	return srv.ListenAndServe()
}
