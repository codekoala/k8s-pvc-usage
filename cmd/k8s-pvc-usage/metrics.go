package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var (
	labels     = []string{"name", "namespace"}
	pvcAvail   *prometheus.GaugeVec
	pvcAvailMB *prometheus.GaugeVec
	pvcUsage   *prometheus.GaugeVec
	pvcUsageMB *prometheus.GaugeVec

	customLabelKeys   []string
	customLabelValues []string
)

func init() {
	customLabelKeys, customLabelValues = readAnnotations()
	labels = append(labels, customLabelKeys...)

	pvcAvail = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "avail",
		Help:      "Percentage of PVC available",
	}, labels)

	pvcAvailMB = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "avail_mb",
		Help:      "Amount of PVC available, in MB",
	}, labels)

	pvcUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "usage",
		Help:      "Percentage of PVC used",
	}, labels)

	pvcUsageMB = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "k8s_pvc",
		Name:      "usage_mb",
		Help:      "Amount of PVC used, in MB",
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
		values = append(values, bits[1])
	}

	return keys, values
}
