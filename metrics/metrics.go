// Package metrics exposes Prometheus metrics for the HTTP API and database pool.
package metrics

import (
	"crypto/subtle"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// unmatchedRoute labels requests that hit no route, keeping label cardinality
// bounded (raw paths of 404s are attacker-controlled).
const unmatchedRoute = "unmatched"

type Metrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

// New creates a private registry with Go runtime, process, HTTP and (if db is
// not nil) database pool collectors.
func New(db *sql.DB) *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, route template and status code.",
		}, []string{"method", "route", "status"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency by method and route template.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being served.",
		}),
	}
	m.registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		m.requests, m.duration, m.inFlight,
	)
	if db != nil {
		m.registry.MustRegister(collectors.NewDBStatsCollector(db, "taskservice"))
	}
	return m
}

// Middleware records request count, latency and concurrency.
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.inFlight.Inc()
		defer m.inFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = unmatchedRoute
		}
		m.requests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.duration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

// Handler serves the metrics. If token is not empty, requests must carry
// "Authorization: Bearer <token>".
func (m *Metrics) Handler(token string) gin.HandlerFunc {
	h := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		if token != "" {
			got := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
			if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(c.Writer, c.Request)
	}
}
