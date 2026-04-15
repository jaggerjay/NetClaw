package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netclaw/proxy-core/internal/api"
	"netclaw/proxy-core/internal/proxy"
	"netclaw/proxy-core/internal/store"
)

func main() {
	cfg := proxy.DefaultConfig()
	st := store.NewMemoryStore()

	proxyServer := proxy.NewServer(cfg, st)
	apiServer := &http.Server{
		Addr:    "127.0.0.1:9091",
		Handler: api.NewServer(st).Handler(),
	}

	go func() {
		log.Printf("proxy listening on %s", cfg.ListenAddress)
		if err := proxyServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("proxy server error: %v", err)
		}
	}()

	go func() {
		log.Printf("api listening on %s", apiServer.Addr)
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("api server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = proxyServer.Close()
	_ = apiServer.Shutdown(ctx)
	log.Println("netclaw stopped")
}
