package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var (
	labels        = []string{"name", "namespace"}
	pvcAvail      *prometheus.GaugeVec
	pvcAvailBytes *prometheus.GaugeVec
	pvcUsage      *prometheus.GaugeVec
	pvcUsageBytes *prometheus.GaugeVec

	customLabelKeys   []string
	customLabelValues []string
)

func init() {
	customLabelKeys, customLabelValues = readAnnotations()
	labels = append(labels, customLabelKeys...)
	log.Info().
		Strs("custom_labels", customLabelKeys).
		Strs("custom_label_values", customLabelValues).
		Msg("using custom labels")

	pvcAvail = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "avail_percent",
		Help:      "Percentage of PVC available",
	}, labels)

	pvcAvailBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "avail_bytes",
		Help:      "Amount of PVC available in bytes",
	}, labels)

	pvcUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "usage_percent",
		Help:      "Percentage of PVC used",
	}, labels)

	pvcUsageBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "usage_bytes",
		Help:      "Amount of PVC used in bytes",
	}, labels)
}

func readAnnotations() (keys, values []string) {
	var (
		line string
		bits []string
	)

	buf, err := os.Open(AnnotationsPath)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", AnnotationsPath).
			Msg("failed to read annotations")
		return
	}
	defer buf.Close()

	scanner := bufio.NewScanner(buf)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line = scanner.Text()
		if !strings.HasPrefix(line, AnnotationsPrefix) || !strings.Contains(line, "=") {
			log.Debug().
				Str("line", line).
				Msg("skipping annotation")
			continue
		}

		line = strings.TrimSpace(strings.TrimPrefix(line, AnnotationsPrefix))
		bits = strings.SplitN(line, "=", 2)

		keys = append(keys, bits[0])
		values = append(values, strings.Trim(bits[1], "\""))
	}

	return keys, values
}
