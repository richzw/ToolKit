package snippet

import (
	"bytes"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	httpRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "The total number of processed http requests",
		},
		[]string{"method", "host", "path", "project", "statusCode"},
	)
	httpRequestDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_request_duration",
		Help: "The duration of processed http requests",
	}, []string{"method", "host", "path", "project", "statusCode"})
)

func nowMillisecond() (now int64) {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type httpMetric struct {
	next http.Handler

	http.ResponseWriter
	statusCode int
}

func (hm *httpMetric) WriteHeader(code int) {
	hm.statusCode = code
	hm.ResponseWriter.WriteHeader(code)
}

func (hm *httpMetric) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	start := nowMillisecond()
	path := req.URL.Path
	defer func() {
		httpRequestDuration.WithLabelValues(req.Method, req.Host, path, z.PodType, strconv.Itoa(hm.statusCode)).Observe((float64)(nowMillisecond() - start))
		httpRequestTotal.WithLabelValues(req.Method, req.Host, path, z.PodType, strconv.Itoa(hm.statusCode)).Inc()
	}()

	hm.next.ServeHTTP(res, req)
}

func NewHttpMetric(handle http.Handler) *httpMetric {
	return &httpMetric{next: handle, statusCode: http.StatusOK}
}

type httpPreRoute struct {
	next http.Handler
}

func (hpr *httpPreRoute) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	reqURI := req.RequestURI
	path := req.URL.Path

	if reqURI == "/" || reqURI == "/v2" {
		res.WriteHeader(http.StatusOK)
		_, err := res.Write([]byte("Commgame server"))
		z.Ast(err)
		return
	}

	if path == "/metrics" || path == "/v2/metrics" {
		if req.Method != "GET" || req.URL.Query().Get("auth") != "qnLcqJZuXAR8Wj0Y" {
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte("Commgame server"))
			if err != nil {
				res.WriteHeader(http.StatusForbidden)
			}
			return
		} else {
			promhttp.Handler().ServeHTTP(res, req)
			return
		}
	}
	hpr.next.ServeHTTP(res, req)
}

func NewHttpPreRoute(handler http.Handler) *httpPreRoute {
	return &httpPreRoute{handler}
}

func ListenAndServe(addr string, mux *runtime.ServeMux) {

	wrapMux := NewHttpMetric(NewHttpPreRoute(mux))
	http.ListenAndServe(addr, wrapMux)
}
