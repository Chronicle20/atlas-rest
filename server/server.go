package server

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"sync"
	"time"
)

type ConfigFunc func(config *Config)

type Config struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	addr         string
}

func NewServer(cl *logrus.Logger, ctx context.Context, wg *sync.WaitGroup, routerProducer func(l logrus.FieldLogger) http.Handler, configurators ...ConfigFunc) {
	l := cl.WithFields(logrus.Fields{"originator": "HTTPServer"})
	w := cl.Writer()
	defer func() {
		err := w.Close()
		if err != nil {
			l.WithError(err).Errorf("Closing log writer.")
		}
	}()

	config := &Config{
		readTimeout:  time.Duration(5) * time.Second,
		writeTimeout: time.Duration(10) * time.Second,
		idleTimeout:  time.Duration(120) * time.Second,
		addr:         ":8080",
	}

	for _, configurator := range configurators {
		configurator(config)
	}

	hs := http.Server{
		Addr:         config.addr,
		Handler:      routerProducer(l),
		ErrorLog:     log.New(w, "", 0),
		ReadTimeout:  config.readTimeout,
		WriteTimeout: config.writeTimeout,
		IdleTimeout:  config.idleTimeout,
	}

	l.Infoln("Starting server on port 8080")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		wg.Add(1)
		defer wg.Done()
		err := hs.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			l.WithError(err).Errorf("Error while serving.")
			return
		}
	}()

	<-ctx.Done()
	l.Infof("Shutting down server on port 8080")
	err := hs.Close()
	if err != nil {
		l.WithError(err).Errorf("Error shutting down HTTP service.")
	}
}

func CommonHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(l logrus.FieldLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Debugf("Handling [%s] request on [%s]", r.Method, r.RequestURI)
			next.ServeHTTP(w, r)
		})
	}
}
