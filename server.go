package socle

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
	"time"
)

func (s *Socle) ListenAndServe() error {

	srv := &http.Server{
		Addr:         s.Server.getURL(),
		ErrorLog:     s.Log.ErrorLog,
		Handler:      s.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	if s.DB.Pool != nil {
		defer s.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}

	if badgerConn != nil {
		defer badgerConn.Close()
	}

	go s.listenRPC()
	s.Log.InfoLog.Printf("Listening on  %s with security %v", s.Server.getURL(), s.Server.Secure)
	if s.Server.Secure {
		s.Log.InfoLog.Println("Begin TLS  Security")
		if s.Server.Security.Strategy == "self" {
			srv.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS13,
			}
			caBytes, err := os.ReadFile(s.Server.Security.CAName + ".crt")
			if err != nil {
				log.Fatal(err)
			}
			ca := x509.NewCertPool()
			if !ca.AppendCertsFromPEM(caBytes) {
				log.Fatal("CA cert not valid")
			}
			srv.TLSConfig.ClientCAs = ca

			if s.Server.Security.MutualTLS {
				srv.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert

			}
		}

		return srv.ListenAndServeTLS(s.Server.Security.ServerCertName+".crt", s.Server.Security.ServerCertName+".key")
	}
	s.Log.InfoLog.Println("Skip TLS  Security")
	return srv.ListenAndServe()
}
