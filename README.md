# k8s-pvc-usage

Barebones PVC usage metrics exposed with Prometheus.

This project is designed to be executed as a single-pod deployment within a Kubernetes cluster.

## Assumptions

- A service account token is available at `/var/run/secrets/kubernetes.io/serviceaccount/token`. The directory can be overridden using the `KUBERNETES_SECRETS_PATH` environment variable.
