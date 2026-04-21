package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"netclaw/proxy-core/internal/api"
	"netclaw/proxy-core/internal/cert"
	"netclaw/proxy-core/internal/proxy"
	"netclaw/proxy-core/internal/store"
)

func main() {
	cfg := proxy.DefaultConfig()
	apiListen := "127.0.0.1:9091"

	flag.StringVar(&cfg.ListenAddress, "proxy-listen", cfg.ListenAddress, "proxy listen address")
	flag.StringVar(&apiListen, "api-listen", apiListen, "API listen address")
	flag.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "data directory for certificates and session database")
	flag.BoolVar(&cfg.CaptureBodies, "capture-bodies", cfg.CaptureBodies, "capture request/response bodies")
	flag.Int64Var(&cfg.MaxBodyBytes, "max-body-bytes", cfg.MaxBodyBytes, "maximum captured bytes per request/response body")
	flag.BoolVar(&cfg.UpstreamTLSSkipVerify, "upstream-tls-skip-verify", cfg.UpstreamTLSSkipVerify, "skip upstream TLS certificate verification")
	flag.Parse()

	st, err := store.NewSQLiteStore(filepath.Join(cfg.DataDir, "sessions.sqlite"))
	if err != nil {
		log.Fatalf("session store setup failed: %v", err)
	}
	defer func() {
		if err := st.Close(); err != nil {
			log.Printf("session store close error: %v", err)
		}
	}()

	authority, err := cert.NewAuthority(filepath.Join(cfg.DataDir, "certs"))
	if err != nil {
		log.Fatalf("certificate authority setup failed: %v", err)
	}

	runtimeInfo := api.RuntimeInfo{
		ProxyListenAddress: cfg.ListenAddress,
		APIListenAddress:   apiListen,
		DataDir:            cfg.DataDir,
		CertificatePath:    authority.Info().CertificatePath,
	}

	proxyServer := proxy.NewServer(cfg, st, authority)
	apiServer := &http.Server{
		Addr:    apiListen,
		Handler: api.NewServer(st, authority, runtimeInfo).Handler(),
	}

	log.Printf("root CA certificate: %s", authority.Info().CertificatePath)
	log.Printf("session database: %s", filepath.Join(cfg.DataDir, "sessions.sqlite"))

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
