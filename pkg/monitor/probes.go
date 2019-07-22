package monitor

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yangyongzhi/sym-operator/pkg/helm"
	"net/http"
)

const MonitorAddr = ":44100"

const (
	isReadyMessage = "I am health."
	isAliveMessage = "I am alive."
	notLiveMessage = "The container is not live now."
)

type LivenessHandler struct {
	helmClient *helm.Client
}

func NewLivenessHandler(helmClient *helm.Client) *LivenessHandler {
	livenessHandler := &LivenessHandler{
		helmClient: helmClient,
	}
	return livenessHandler
}

func (l *LivenessHandler) livenessProbe(w http.ResponseWriter, r *http.Request) {
	err := l.helmClient.KeepLive()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(notLiveMessage))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(isAliveMessage))
	}
}

func readinessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(isReadyMessage))
}

func NewProbesMux(helmClient *helm.Client) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/readiness", readinessProbe)
	mux.HandleFunc("/liveness", NewLivenessHandler(helmClient).livenessProbe)
	return mux
}

func AddPrometheusHandler(mux *http.ServeMux) {
	// Register HTTP handler for the global Prometheus registry.
	mux.Handle("/metrics", promhttp.Handler())
}
