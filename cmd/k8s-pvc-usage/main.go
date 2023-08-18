package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"diegomarangoni.dev/typenv"
	pvcusage "github.com/codekoala/k8s-pvc-usage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	BindAddr       = typenv.String("BIND_ADDR", ":9100")
	ScrapeInterval = typenv.Duration("SCRAPE_INTERVAL", 15*time.Second)

	SecretsPath = typenv.String("KUBERNETES_SECRETS_PATH", "/var/run/secrets/kubernetes.io/serviceaccount")
	ApiHost     = typenv.String("KUBERNETES_SERVICE_HOST", "kubernetes.default")
	ApiPort     = typenv.String("KUBERNETES_PORT_443_TCP_PORT", "443")
	ApiTimeout  = typenv.Duration("API_REQUEST_TIMEOUT", 5*time.Second)

	pvcUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "usage",
		Help:      "Percentage of PVC used",
	}, []string{"name", "namespace"})

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

	prometheus.MustRegister(pvcUsage)

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
			pvcUsage.WithLabelValues(pvc.Name, pvc.Namespace).Set(pvc.Usage())
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
