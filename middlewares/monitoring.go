package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 1.5, 2, 3, 5},
		},
		[]string{"method", "path", "status"},
	)

	apiErrorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_error_count",
			Help: "Total number of API errors by endpoint and status",
		},
		[]string{"endpoint", "status"},
	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path // 使用原始路径如果未匹配路由
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// 记录通用指标
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Observe(duration)

		// 记录错误指标（4xx和5xx）
		if c.Writer.Status() >= http.StatusBadRequest {
			apiErrorCount.WithLabelValues(
				path,
				status,
			).Inc()
		}
	}
}
