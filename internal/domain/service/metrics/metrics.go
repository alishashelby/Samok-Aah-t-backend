package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	RequestsInFlight prometheus.Gauge

	ClientsTotal    prometheus.Gauge
	ModelsTotal     prometheus.Gauge
	CompletedOrders prometheus.Counter
}

func NewMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "app_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "app_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		RequestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "app_http_requests_in_flight",
			Help: "Number of HTTP requests currently in flight",
		}),
		ClientsTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "app_clients_total",
			Help: "Total number of registered clients",
		}),
		ModelsTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "app_models_total",
			Help: "Total number of registered models",
		}),
		CompletedOrders: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "app_completed_orders_total",
			Help: "Total number of completed orders",
		}),
	}

	prometheus.MustRegister(
		m.RequestsTotal,
		m.RequestDuration,
		m.RequestsInFlight,
		m.ClientsTotal,
		m.ModelsTotal,
		m.CompletedOrders,
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

func (m *Metrics) ObserveRequest(method string, status string, duration time.Duration) {
	m.RequestsTotal.WithLabelValues(method, status).Inc()
	m.RequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

func (m *Metrics) SetClients(count int64) {
	m.ClientsTotal.Set(float64(count))
}

func (m *Metrics) SetModels(count int64) {
	m.ModelsTotal.Set(float64(count))
}

func (m *Metrics) IncCompletedOrders() {
	m.CompletedOrders.Inc()
}
