package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	pvcusage "github.com/codekoala/k8s-pvc-usage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	version = "0.0.0"
	branch  = ""
	commit  = ""
	date    = ""
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	defer close(sig)

	signal.Notify(sig, os.Kill, os.Interrupt)
	go func() {
		defer cancel()
		<-sig
	}()

	prometheus.MustRegister(pvcAvail)
	prometheus.MustRegister(pvcAvailMB)
	prometheus.MustRegister(pvcUsage)
	prometheus.MustRegister(pvcUsageMB)

	tokenPath := filepath.Join(SecretsPath, "token")
	token, err := os.ReadFile(tokenPath)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("path", tokenPath).
			Msg("unable to read API token")
	}

	api := pvcusage.New(net.JoinHostPort(ApiHost, ApiPort), string(token))
	api.With(
		pvcusage.CA(SecretsPath),
		pvcusage.Timeout(ApiTimeout),
	)

	log.Info().Msg("client configured")

	go scrapeUsage(ctx, api)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(BindAddr, nil); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}

func scrapeUsage(ctx context.Context, api *pvcusage.Client) {
	var counter uint

	ticker := time.NewTicker(ScrapeInterval)
	defer ticker.Stop()

	for {
		log.Debug().Msg("scraping stats")
		counter = 0

		for _, pvc := range pvcusage.GetPvcUsageCtx(ctx, api) {
			labels := append([]string{pvc.Name, pvc.Namespace}, customLabelValues...)
			pvcAvail.WithLabelValues(labels...).Set(pvc.Avail())
			pvcAvailMB.WithLabelValues(labels...).Set(pvc.AvailableBytes / 1048576)
			pvcUsage.WithLabelValues(labels...).Set(pvc.Usage())
			pvcUsageMB.WithLabelValues(labels...).Set(pvc.UsedBytes / 1048576)
			counter++
		}

		log.Info().Uint("count", counter).Msg("scraped metrics for PVCs")
		select {
		case <-ctx.Done():
			log.Info().Msg("exiting scrape loop")
			return
		case <-ticker.C:
			continue
		}
	}
}
