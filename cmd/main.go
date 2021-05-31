package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/sirupsen/logrus"

	cmdConfig "github.com/k8s-at-home/gateway-admision-controller/internal/config"
	"github.com/k8s-at-home/gateway-admision-controller/internal/http/webhook"
	"github.com/k8s-at-home/gateway-admision-controller/internal/log"
)

type config struct {
	certFile string
	keyFile  string
	gateway  string
}

func runApp() error {

	cfg, err := cmdConfig.NewCmdConfig()
	if err != nil {
		return fmt.Errorf("could not get commandline configuration: %w", err)
	}

	// Set up logger.
	logrusLog := logrus.New()
	logrusLogEntry := logrus.NewEntry(logrusLog).WithField("app", "gateway-admision-controller")
	if cfg.Debug {
		logrusLogEntry.Logger.SetLevel(logrus.DebugLevel)
	}
	if !cfg.Development {
		logrusLogEntry.Logger.SetFormatter(&logrus.JSONFormatter{})
	}
	logger := log.NewLogrus(logrusLogEntry).WithKV(log.KV{"version": cmdConfig.Version})

	// Prepare run entrypoints.
	var g run.Group

	// OS signals.
	{
		sigC := make(chan os.Signal, 1)
		exitC := make(chan struct{})
		signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)

		g.Add(
			func() error {
				select {
				case s := <-sigC:
					logger.Infof("signal %s received", s)
					return nil
				case <-exitC:
					return nil
				}
			},
			func(_ error) {
				close(exitC)
			},
		)
	}

	// Webhook HTTP server.
	{
		logger := logger.WithKV(log.KV{"addr": cfg.WebhookListenAddr, "http-server": "webhooks"})

		// Webhook handler.
		wh, err := webhook.New(webhook.Config{
			CmdConfig: *cfg,
			Logger:    logger,
		})
		if err != nil {
			return fmt.Errorf("could not create webhooks handler: %w", err)
		}

		mux := http.NewServeMux()
		mux.Handle("/", wh)
		server := http.Server{Addr: cfg.WebhookListenAddr, Handler: mux}

		g.Add(
			func() error {
				if cfg.TLSCertFilePath == "" || cfg.TLSKeyFilePath == "" {
					logger.Warningf("webhook running without TLS")
					logger.Infof("http server listening...")
					return server.ListenAndServe()
				}

				logger.Infof("https server listening...")
				return server.ListenAndServeTLS(cfg.TLSCertFilePath, cfg.TLSKeyFilePath)
			},
			func(_ error) {
				logger.Infof("start draining connections")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				err := server.Shutdown(ctx)
				if err != nil {
					logger.Errorf("error while shutting down the server: %s", err)
				} else {
					logger.Infof("server stopped")
				}
			},
		)
	}

	err = g.Run()

	return err
}

func main() {
	err := runApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running app: %s", err)
		os.Exit(1)
	}

	os.Exit(0)
}
