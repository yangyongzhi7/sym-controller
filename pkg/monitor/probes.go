package monitor

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const MonitorAddr = ":44100"

const (
	isReadyMessage = "I am health."
	isAliveMessage = "I am alive."
)

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(isReadyMessage))
}

func livenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(isAliveMessage))
}

func NewProbesMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", readinessProbe)
	mux.HandleFunc("/liveness", livenessProbe)
	return mux
}

func AddPrometheusHandler(mux *http.ServeMux) {
	// Register HTTP handler for the global Prometheus registry.
	mux.Handle("/metrics", promhttp.Handler())
}
