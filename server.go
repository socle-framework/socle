package socle

import (
	"log"
	"net/http"
	"os"
	"time"
)

func (c *Socle) ListenAndServe(addr string) error {
	log.Printf("App start with url  %v", addr)
	srv := &http.Server{
		Addr:         addr,
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
