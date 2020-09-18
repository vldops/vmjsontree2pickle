package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (
	prometheusRequestsTotalCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requestsTotal",
			Help: "requestsTotal",
		},
		[]string{"method", "path", "statusCode"},
	)

	prometheusRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "requestsDuration",
			Help:    "requestsDuration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "statusCode"},
	)
)

type reQuest struct {
	http.ResponseWriter
	handler              string
	method               string
	path                 string
	statusCode           int
	statusCodePrometheus string
	startedAt            time.Time
	finishedAt           time.Time
	duration             float64
	remoteAddr           string
}

func makeReQuest(w http.ResponseWriter) *reQuest {
	return &reQuest{
		ResponseWriter: w,
	}
}

// WriteHeader method
func (r *reQuest) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
	r.statusCodePrometheus = strconv.Itoa(statusCode)
}

func (r *reQuest) started() {
	r.startedAt = time.Now()
}

func (r *reQuest) finished() {
	r.finishedAt = time.Now()
	r.duration = time.Since(r.startedAt).Seconds()
}

func (r *reQuest) metrics(path string, method string, remoteAddr string) {
	r.path = path
	r.method = method
	r.remoteAddr = remoteAddr
}

func loggerMiddleware() func(http.Handler) http.Handler {
	return func(serve http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			request := makeReQuest(w)
			request.started()

			serve.ServeHTTP(request, r)

			request.finished()
			request.metrics(r.URL.EscapedPath(), r.Method, strings.Split(r.RemoteAddr, ":")[0])

			prometheusRequestsTotalCounter.With(prometheus.Labels{
				"method":     request.method,
				"path":       request.path,
				"statusCode": request.statusCodePrometheus,
			}).Inc()

			prometheusRequestsDuration.With(prometheus.Labels{
				"method":     request.method,
				"path":       request.path,
				"statusCode": request.statusCodePrometheus,
			}).Observe(request.duration)

			logger.Info("metricsMessage",
				zap.Int("statusCode", request.statusCode),
				zap.String("method", request.method),
				zap.String("path", request.path),
				zap.Float64("requestTime", request.duration),
				zap.String("remoteAddr", request.remoteAddr),
				zap.Any("queryArgs", r.URL.Query()),
			)

		})
	}
}
