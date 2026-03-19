package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)

	// Analysis metrics
	AnalysisTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analysis_total",
			Help: "Total number of page analyses",
		},
		[]string{"status"},
	)

	AnalysisDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analysis_duration_seconds",
			Help:    "Time taken to analyze a page",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
	)

	LinksValidated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "links_validated_total",
			Help: "Total number of links validated",
		},
		[]string{"type", "status"},
	)

	// Rate limiting metrics
	RateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of rate limit violations",
		},
		[]string{"ip"},
	)

	ActiveVisitors = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_visitors",
			Help: "Current number of active visitors",
		},
	)

	// Error metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "operation"},
	)
)

// MetricsMiddleware instruments HTTP handlers with Prometheus metrics
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		HttpRequestsInFlight.Inc()
		defer HttpRequestsInFlight.Dec()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		HttpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		HttpRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecordAnalysis records analysis metrics
func RecordAnalysis(duration time.Duration, success bool, internal, external, inaccessible int) {
	AnalysisDuration.Observe(duration.Seconds())

	status := "success"
	if !success {
		status = "failure"
	}
	AnalysisTotal.WithLabelValues(status).Inc()

	LinksValidated.WithLabelValues("internal", "checked").Add(float64(internal))
	LinksValidated.WithLabelValues("external", "checked").Add(float64(external))
	LinksValidated.WithLabelValues("any", "inaccessible").Add(float64(inaccessible))
}

// RecordRateLimitExceeded records rate limit violation
func RecordRateLimitExceeded(ip string) {
	RateLimitExceeded.WithLabelValues(ip).Inc()
}

// UpdateActiveVisitors updates the active visitors gauge
func UpdateActiveVisitors(count int) {
	ActiveVisitors.Set(float64(count))
}

// RecordError records an error occurrence
func RecordError(errorType, operation string) {
	ErrorsTotal.WithLabelValues(errorType, operation).Inc()
}