package metric

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusFactory interface {
	InitHandler() http.Handler
}

type prometheusFactory struct {
	registry *prometheus.Registry
}

func NewPrometheusFactory(cs ...prometheus.Collector) PrometheusFactory {
	f := &prometheusFactory{}
	f.registry = f.initRegistry(cs)

	return f
}

func (f *prometheusFactory) InitHandler() http.Handler {
	return promhttp.HandlerFor(f.registry,
		promhttp.HandlerOpts{
			Registry:          f.registry,
			EnableOpenMetrics: true,
		})
}

func (f *prometheusFactory) initRegistry(cs []prometheus.Collector) *prometheus.Registry {
	reg := prometheus.NewRegistry()

	defaultCs := []prometheus.Collector{
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	}

	fCS := make([]prometheus.Collector, 0, len(defaultCs)+len(cs))
	fCS = append(fCS, defaultCs...)

	for _, c := range cs {
		if c != nil {
			fCS = append(fCS, c)
		}
	}

	reg.MustRegister(fCS...)

	return reg
}
