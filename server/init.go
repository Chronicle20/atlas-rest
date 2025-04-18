package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

type RouteInitializer func(*mux.Router, logrus.FieldLogger)

// CreateService Deprecated
//
//goland:noinspection GoUnusedExportedFunction
func CreateService(l *logrus.Logger, ctx context.Context, wg *sync.WaitGroup, basePath string, initializers ...RouteInitializer) {
	New(l).
		WithContext(ctx).
		WithWaitGroup(wg).
		SetBasePath(basePath).
		SetRouteInitializers(initializers...).
		Run()
}

// ProduceRoutes Deprecated
func ProduceRoutes(basePath string, initializers ...RouteInitializer) func(l logrus.FieldLogger) http.Handler {
	return func(l logrus.FieldLogger) http.Handler {
		router := mux.NewRouter().PathPrefix(basePath).Subrouter().StrictSlash(true)
		router.Use(CommonHeader)
		router.Use(LoggingMiddleware(l))

		for _, initializer := range initializers {
			initializer(router, l)
		}

		return router
	}
}
