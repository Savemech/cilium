package metrics

// metrics holds prometheus metrics objects and related utility functions. It
// does not abstract away the prometheus client but the caller rarely needs to
// refer to prometheus directly.
//
// **Adding a metric**
// - Add a metric object of the appropriate type that is exported
// - Register the new object in the init function

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	registry = prometheus.NewRegistry()

	// Namespace is used to scope metrics from cilium. It is prepended to metric
	// names and separated with a '_'
	Namespace = "cilium"

	// Endpoint

	// EndpointCount is the number of managed endpoints
	EndpointCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "endpoint_count",
		Help:      "Number of endpoints managed by this agent",
	})

	// EndpointCountRegenerating is the number of endpoints currently regenerating
	EndpointCountRegenerating = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "endpoint_regenerating",
		Help:      "Number of endpoints currently regenerating",
	})

	// CountEndpointsRegenerations is a count of the number of times any endpoint
	// has been regenerated and success/fail outcome
	CountEndpointsRegenerations = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "endpoint_regenerations",
		Help:      "Count of all endpoint regenerations that have completed, tagged by outcome",
	},
		[]string{"outcome"})

	// Policy

	// PolicyCount is the number of policies loaded into the agent
	PolicyCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "policy_count",
		Help:      "Number of policies currently loaded",
	})

	// PolicyRevision is the current policy revision number for this agent
	PolicyRevision = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "policy_max_revision",
		Help:      "Highest policy revision number in the agent",
	})

	// PolicyImportErrors is a count of failed policy imports
	PolicyImportErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "policy_import_errors",
		Help:      "Number of times a policy import has failed",
	})
)

func init() {
	registry.MustRegister(prometheus.NewProcessCollector(os.Getpid(), Namespace))
	// TODO: Figure out how to put this into a Namespace
	//registry.MustRegister(prometheus.NewGoCollector())

	registry.MustRegister(EndpointCount)
	registry.MustRegister(EndpointCountRegenerating)
	registry.MustRegister(CountEndpointsRegenerations)

	registry.MustRegister(PolicyCount)
	registry.MustRegister(PolicyRevision)
	registry.MustRegister(PolicyImportErrors)
}

// Enable begins serving prometheus metrics on the address passed in. Addresses
// of the form ":8080" will bind the port on all interfaces.
func Enable(addr string) error {
	go func() {
		// The Handler function provides a default handler to expose metrics
		// via an HTTP server. "/metrics" is the usual endpoint for that.
		http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		log.WithError(http.ListenAndServe(addr, nil)).Warn("Cannot start metrics server on %s", addr)
	}()

	return nil
}
