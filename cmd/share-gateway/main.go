package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"example.com/temporary-share-gateway/internal/clock"
	"example.com/temporary-share-gateway/internal/config"
	"example.com/temporary-share-gateway/internal/gateway"
	"example.com/temporary-share-gateway/internal/metrics"
	"example.com/temporary-share-gateway/internal/persist"
	"example.com/temporary-share-gateway/internal/security"
	"example.com/temporary-share-gateway/internal/share"
	"example.com/temporary-share-gateway/internal/transport"
)

func main() {
	settings, err := config.Environment()
	if err != nil {
		log.Fatal(err)
	}
	store, err := persist.Open(settings.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	_ = store.SaveConfig(settings.Gateway)
	resources := gateway.StaticResources{"demo-resource": "temporary share gateway is ready"}
	fixedClock := clock.NewFixed(clock.Unix(1700000000))
	service := share.NewService(store, fixedClock, resources)
	auditor := security.NewAuditor(store)
	registry := share.NewRegistry(service, share.DefaultPolicy())
	counters := metrics.New()
	handler := gateway.NewHandler(service, auditor, resources, counters, settings.Gateway.RequireRequestID)
	mux := httpMux(handler, gateway.NewAdminHandler(registry, store, fixedClock.Now), store, counters)
	server := transport.NewServer(settings.ListenAddress, mux)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if serveErr := server.ListenAndServe(); serveErr != nil {
			log.Print(serveErr)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), settings.Gateway.RequestTimeout)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func httpMux(handler *gateway.Handler, admin *gateway.AdminHandler, store *persist.Store, counters *metrics.Counter) http.Handler {
	router := transport.NewRouter(counters)
	router.MountShare(handler)
	router.MountAdmin(admin)
	router.MountHealth(store.ValidateReady)
	router.MountMetrics()
	return router
}
