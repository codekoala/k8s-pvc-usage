package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/bep/debounce"
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
	prometheus.MustRegister(pvcAvailBytes)
	prometheus.MustRegister(pvcUsage)
	prometheus.MustRegister(pvcUsageBytes)

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

	// avoid resetting metrics in the middle of scraping
	mut := new(sync.Mutex)

	// refresh can be used to force a refresh of the metrics after a reset
	refresh := make(chan bool, 1)
	defer close(refresh)

	go scrapeUsage(ctx, api, mut, refresh)
	go resetUsage(ctx, api, mut, refresh)

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(BindAddr, nil); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}

func scrapeUsage(ctx context.Context, api *pvcusage.Client, mut *sync.Mutex, refresh chan bool) {
	var counter uint

	ticker := time.NewTicker(ScrapeInterval)
	defer ticker.Stop()

	debounced := debounce.New(DebounceDelay)

	scrape := func() {
		// acquire lock on metrics
		mut.Lock()

		// release lock on metrics when we're done here
		defer mut.Unlock()

		log.Debug().Msg("scraping stats")
		counter = 0

		for _, pvc := range pvcusage.GetPvcUsageCtx(ctx, api) {
			labels := append([]string{pvc.Name, pvc.Namespace}, customLabelValues...)
			pvcAvail.WithLabelValues(labels...).Set(pvc.Avail())
			pvcAvailBytes.WithLabelValues(labels...).Set(pvc.AvailableBytes)
			pvcUsage.WithLabelValues(labels...).Set(pvc.Usage())
			pvcUsageBytes.WithLabelValues(labels...).Set(pvc.UsedBytes)
			counter++
		}

		log.Info().Uint("count", counter).Msg("scraped metrics for PVCs")
	}

	for {
		debounced(scrape)

		select {
		case <-ctx.Done():
			log.Info().Msg("exiting scrape loop")
			return
		case <-ticker.C:
			continue
		case <-refresh:
			log.Info().Msg("refresh requested")
			ticker.Reset(ScrapeInterval)
			continue
		}
	}
}

// resetUsage periodically resets all metrics to avoid returning stale information about, for example, PVCs that have been deleted
func resetUsage(ctx context.Context, api *pvcusage.Client, mut *sync.Mutex, refresh chan bool) {
	ticker := time.NewTicker(ResetInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("exiting reset loop")
			return
		case <-ticker.C:
			func() {
				// acquire lock on metrics
				mut.Lock()

				// release lock on metrics when we're done here
				defer mut.Unlock()

				log.Info().Msg("resetting stats")
				pvcAvail.Reset()
				pvcAvailBytes.Reset()
				pvcUsage.Reset()
				pvcUsageBytes.Reset()
			}()

			// immediately scrape usage to avoid missing data
			refresh <- true
		}
	}
}
