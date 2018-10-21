package grpc_prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace      = "grpc"
	labelService   = "grpc_service"
	labelMethod    = "grpc_method"
	labelType      = "grpc_type"
	labelCode      = "grpc_code"
	labelUserAgent = "grpc_user_agent"
	labelFailFast  = "grpc_fail_fast"
)

type monitor struct {
	// Counters
	requestsTotal    *prometheus.CounterVec
	responsesTotal   *prometheus.CounterVec
	messagesReceived *prometheus.CounterVec
	messagesSent     *prometheus.CounterVec
	errors           *prometheus.CounterVec
	// Gauges
	connections      *prometheus.GaugeVec
	inFlightRequests *prometheus.GaugeVec
	// Histograms
	requestDuration *prometheus.HistogramVec
}

var _ prometheus.Collector = &monitor{}

// Describe implements prometheus Collector interface.
func (m *monitor) Describe(in chan<- *prometheus.Desc) {
	// CounterVec
	m.requestsTotal.Describe(in)
	m.responsesTotal.Describe(in)
	m.messagesSent.Describe(in)
	m.messagesReceived.Describe(in)
	m.errors.Describe(in)

	// Gauge
	m.connections.Describe(in)
	m.inFlightRequests.Describe(in)

	// HistogramVec
	m.requestDuration.Describe(in)
}

// Collect implements prometheus Collector interface.
func (m *monitor) Collect(in chan<- prometheus.Metric) {
	// CounterVec
	m.requestsTotal.Collect(in)
	m.responsesTotal.Collect(in)
	m.messagesSent.Collect(in)
	m.messagesReceived.Collect(in)
	m.errors.Collect(in)

	// Gauge
	m.connections.Collect(in)
	m.inFlightRequests.Collect(in)

	// HistogramVec
	m.requestDuration.Collect(in)

}

type clientMonitor struct {
	dialer *prometheus.CounterVec
	*monitor
}

var _ prometheus.Collector = &clientMonitor{}

func initClientMonitor(opts prometheus.Opts) *clientMonitor {
	dialer := prometheus.NewCounterVec(
		counterOpts(opts, "reconnects_total", "A total number of reconnects made by the client."),
		[]string{"address"},
	)

	// Counters
	requestsTotal := prometheus.NewCounterVec(
		counterOpts(opts, "requests_total", "A total number of reconnects made by the client.."),
		[]string{labelService, labelMethod, labelType},
	)
	responsesTotal := prometheus.NewCounterVec(
		counterOpts(opts, "responses_total", "A total number of RPC responses received by the client."),
		[]string{labelService, labelMethod, labelCode, labelType},
	)
	receivedMessages := prometheus.NewCounterVec(
		counterOpts(opts, "received_messages_total", "A total number of RPC messages received by the client."),
		[]string{labelService, labelMethod, labelType},
	)
	sentMessages := prometheus.NewCounterVec(
		counterOpts(opts, "sent_messages_total", "A total number of RPC messages sent by the client."),
		[]string{labelService, labelMethod, labelType},
	)
	errors := prometheus.NewCounterVec(
		counterOpts(opts, "errors_total", "A total number of errors that happen during RPC calls."),
		[]string{labelService, labelMethod, labelCode, labelType},
	)

	// Gauges
	connections := prometheus.NewGaugeVec(
		gaugeOpts(opts, "connections", "A current number of incoming RPC connections."),
		[]string{"remote_addr", "local_addr"},
	)
	inFlightRequests := prometheus.NewGaugeVec(
		gaugeOpts(opts, "in_flight_requests", "A number of currently processed client-side RPC requests."),
		[]string{labelFailFast, labelMethod, labelService},
	)

	// Histograms
	requestDuration := prometheus.NewHistogramVec(
		histogramOpts(opts, "request_duration_histogram_seconds", "The RPC request latencies in seconds on the client side.", nil),
		[]string{labelService, labelMethod, labelCode, labelType},
	)

	return &clientMonitor{
		dialer: dialer,
		monitor: &monitor{
			connections:      connections,
			inFlightRequests: inFlightRequests,
			requestsTotal:    requestsTotal,
			responsesTotal:   responsesTotal,
			requestDuration:  requestDuration,
			messagesReceived: receivedMessages,
			messagesSent:     sentMessages,
			errors:           errors,
		},
	}
}

// Describe implements prometheus Collector interface.
func (cm *clientMonitor) Describe(in chan<- *prometheus.Desc) {
	cm.dialer.Describe(in)
	cm.monitor.Describe(in)
}

// Collect implements prometheus Collector interface.
func (cm *clientMonitor) Collect(in chan<- prometheus.Metric) {
	cm.dialer.Collect(in)
	cm.monitor.Collect(in)
}

type serverMonitor struct {
	*monitor
}

var _ prometheus.Collector = &clientMonitor{}

func initServerMonitor(opts prometheus.Opts) *serverMonitor {
	baseLabels := []string{labelService, labelMethod, labelUserAgent, labelType}

	// Counters
	requestsTotal := prometheus.NewCounterVec(
		counterOpts(opts, "received_requests_total", "A total number of RPC requests received by the server."),
		labelsWith(baseLabels),
	)
	responsesTotal := prometheus.NewCounterVec(
		counterOpts(opts, "responses_total", "A total number of RPC responses sent back by the server."),
		labelsWith(baseLabels, labelCode),
	)
	receivedMessages := prometheus.NewCounterVec(
		counterOpts(opts, "received_messages_total", "A total number of RPC messages received by the server."),
		baseLabels,
	)
	sentMessages := prometheus.NewCounterVec(
		counterOpts(opts, "sent_messages_total", "A total number of RPC messages sent by the server."),
		baseLabels,
	)
	errors := prometheus.NewCounterVec(
		counterOpts(opts, "errors_total", "A total number of errors that happen during RPC calls on the server side."),
		labelsWith(baseLabels, labelCode),
	)

	// Gauges
	connections := prometheus.NewGaugeVec(
		gaugeOpts(opts, "connections", "A current number of outgoing RPC connections."),
		[]string{"remote_addr", "local_addr", labelUserAgent},
	)
	inFlightRequests := prometheus.NewGaugeVec(
		gaugeOpts(opts, "in_flight_requests", "A number of currently processed server-side RPC requests."),
		labelsWith(baseLabels, labelFailFast),
	)

	// Histograms
	requestDuration := prometheus.NewHistogramVec(
		histogramOpts(opts, "request_duration_histogram_seconds", "The RPC request latencies in seconds on the server side.", nil),
		labelsWith(baseLabels, labelCode),
	)

	return &serverMonitor{
		monitor: &monitor{
			connections:      connections,
			inFlightRequests: inFlightRequests,
			requestsTotal:    requestsTotal,
			responsesTotal:   responsesTotal,
			requestDuration:  requestDuration,
			messagesReceived: receivedMessages,
			messagesSent:     sentMessages,
			errors:           errors,
		},
	}
}

func histogramOpts(opts prometheus.Opts, name, help string, buckets []float64) prometheus.HistogramOpts {
	opts.Name = name
	opts.Help = help

	return prometheus.HistogramOpts{
		Subsystem:   opts.Subsystem,
		Namespace:   opts.Namespace,
		Name:        name,
		Help:        opts.Help,
		ConstLabels: opts.ConstLabels,
		Buckets:     buckets,
	}
}

func counterOpts(opts prometheus.Opts, name, help string) prometheus.CounterOpts {
	opts.Name = name
	opts.Help = help

	return prometheus.CounterOpts(opts)
}

func gaugeOpts(opts prometheus.Opts, name, help string) prometheus.GaugeOpts {
	opts.Name = name
	opts.Help = help

	return prometheus.GaugeOpts(opts)
}

func labelsWith(base []string, additional ...string) []string {
	res := make([]string, len(base)+len(additional))
	copy(res, base)

	return append(res, additional...)
}
