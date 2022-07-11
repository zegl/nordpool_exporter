package main

import (
	"flag"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "nordpool"

var (
	addr = flag.String("addr", ":9367", "The address to listen on for HTTP requests.")
)

func main() {
	logger, _ := zap.NewProduction()
	logger.Info("Starting nordpool_exporter")

	prometheus.MustRegister(NewNordpoolCollector(namespace, logger))

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Nordpool Exporter</title></head>
            <body>
            <h1>Nordpool Exporter</h1>
            <p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})
	srv := &http.Server{
		Addr:         *addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Listening on", zap.Stringp("addr", addr))
	logger.Fatal("failed to start server", zap.Error(srv.ListenAndServe()))
}
