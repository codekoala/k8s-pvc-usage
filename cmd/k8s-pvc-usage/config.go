package main

import (
	"time"

	"diegomarangoni.dev/typenv"
)

var (
	BindAddr       = typenv.String("BIND_ADDR", ":9100")
	ScrapeInterval = typenv.Duration("SCRAPE_INTERVAL", 15*time.Second)
	ResetInterval  = typenv.Duration("RESET_INTERVAL", 5*time.Minute)

	SecretsPath = typenv.String("KUBERNETES_SECRETS_PATH", "/var/run/secrets/kubernetes.io/serviceaccount")
	ApiHost     = typenv.String("KUBERNETES_SERVICE_HOST", "kubernetes.default")
	ApiPort     = typenv.String("KUBERNETES_PORT_443_TCP_PORT", "443")
	ApiTimeout  = typenv.Duration("API_REQUEST_TIMEOUT", 5*time.Second)

	AnnotationsPath   = typenv.String("POD_ANNOTATION_PATH", "/etc/podinfo/annotations")
	AnnotationsPrefix = typenv.String("POD_ANNOTATION_PREFIX", "k8s-pvc-usage/")
)
