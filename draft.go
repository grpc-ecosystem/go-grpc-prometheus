package grpc_prometheus

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
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

type ctxKey int

var (
	tagRPCKey  ctxKey = 1
	tagConnKey ctxKey = 2
)

type Subsystem int

var (
	Server Subsystem = 1
	Client Subsystem = 2
)

func (s Subsystem) String() string {
	switch s {
	case Server:
		return "server"
	case Client:
		return "client"
	default:
		return "unknown"
	}
}

// NewRequestsTotalGaugeVecV1 exists for backward compatibility.
func NewRequestsTotalCounterVecV1(sub Subsystem) *prometheus.CounterVec {
	name := "started_total"
	help := fmt.Sprintf("Total number of RPCs started on the %s.", sub.String())
	switch sub {
	case Server:
		return newRequestsTotalCounterVec(sub.String(), name, help)
	case Client:
		return newRequestsTotalCounterVec(sub.String(), name, help)
	default:
		// TODO: panic?
		panic("unknown subsystem")
	}
}

func NewRequestsTotalCounterVec(sub Subsystem) *prometheus.CounterVec {
	switch sub {
	case Server:
		return newRequestsTotalCounterVec(sub.String(), "received_requests_total", "A total number of RPC requests received by the server.")
	case Client:
		return newRequestsTotalCounterVec(sub.String(), "sent_requests_total", "A total number of RPC requests sent by the client.")
	default:
		// TODO: panic?
		panic("unknown subsystem")
	}
}

func newRequestsTotalCounterVec(sub, name, help string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: sub,
			Name:      name,
			Help:      help,
		},
		[]string{labelFailFast, labelService, labelMethod},
		//[]string{labelType, labelService, labelMethod}, TODO: IsServerStream and IsClientStream not available outside interceptors. Type label cannot be used.
	)
}

type StatsHandlerCollector interface {
	// Init reallocates possible dimensions for given metric.
	Init(map[string]grpc.ServiceInfo) error

	stats.Handler
	prometheus.Collector
}

type StatsHandler struct {
	handlers []StatsHandlerCollector
}

var _ StatsHandlerCollector = &StatsHandler{}

// NewStatsHandler allows to pass various number of handlers.
func NewStatsHandler(handlers ...StatsHandlerCollector) *StatsHandler {
	return &StatsHandler{
		handlers: handlers,
	}
}

// Init implements StatsHandlerCollector interface.
// TODO: implement
func (h *StatsHandler) Init(info map[string]grpc.ServiceInfo) error {
	return nil
}

func (h *StatsHandler) TagRPC(ctx context.Context, inf *stats.RPCTagInfo) context.Context {
	service, method := split(inf.FullMethodName)

	ctx = context.WithValue(ctx, tagRPCKey, prometheus.Labels{
		labelFailFast: strconv.FormatBool(inf.FailFast),
		labelService:  service,
		labelMethod:   method,
	})

	for _, c := range h.handlers {
		ctx = c.TagRPC(ctx, inf)
	}
	return ctx
}

// HandleRPC processes the RPC stats.
func (h *StatsHandler) HandleRPC(ctx context.Context, sts stats.RPCStats) {
	for _, c := range h.handlers {
		c.HandleRPC(ctx, sts)
	}
}

func (h *StatsHandler) TagConn(ctx context.Context, inf *stats.ConnTagInfo) context.Context {
	for _, c := range h.handlers {
		ctx = c.TagConn(ctx, inf)
	}
	return ctx
}

// HandleConn processes the Conn stats.
func (h *StatsHandler) HandleConn(ctx context.Context, sts stats.ConnStats) {
	for _, c := range h.handlers {
		c.HandleConn(ctx, sts)
	}
}

// Describe implements prometheus Collector interface.
func (h *StatsHandler) Describe(in chan<- *prometheus.Desc) {
	for _, c := range h.handlers {
		c.Describe(in)
	}
}

// Collect implements prometheus Collector interface.
func (h *StatsHandler) Collect(in chan<- prometheus.Metric) {
	for _, c := range h.handlers {
		c.Collect(in)
	}
}

type RequestsTotalStatsHandler struct {
	sub Subsystem
	vec *prometheus.CounterVec
}

// NewRequestsTotalStatsHandler ...
// The GaugeVec must have zero, one, two, three or four non-const non-curried labels.
// For those, the only allowed label names are "fail_fast", "handler", "service" and "user_agent".
func NewRequestsTotalStatsHandler(sub Subsystem, vec *prometheus.CounterVec) *RequestsTotalStatsHandler {
	return &RequestsTotalStatsHandler{
		sub: sub,
		vec: vec,
	}
}

// Init implements StatsHandlerCollector interface.
// TODO: implement
func (h *RequestsTotalStatsHandler) Init(info map[string]grpc.ServiceInfo) error {
	return nil
}

func (h *RequestsTotalStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC processes the RPC stats.
func (h *RequestsTotalStatsHandler) HandleRPC(ctx context.Context, stat stats.RPCStats) {
	lab, _ := ctx.Value(tagRPCKey).(prometheus.Labels)

	if beg, ok := stat.(*stats.Begin); ok {
		switch {
		case beg.IsClient() && h.sub == Client:
			h.vec.With(lab).Inc()
		case !beg.IsClient() && h.sub == Server:
			h.vec.With(lab).Inc()
		}
	}
}

func (h *RequestsTotalStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes the Conn stats.
func (h *RequestsTotalStatsHandler) HandleConn(ctx context.Context, stat stats.ConnStats) {
}

// Describe implements prometheus Collector interface.
func (h *RequestsTotalStatsHandler) Describe(in chan<- *prometheus.Desc) {
	h.vec.Describe(in)
}

// Collect implements prometheus Collector interface.
func (h *RequestsTotalStatsHandler) Collect(in chan<- prometheus.Metric) {
	h.vec.Collect(in)
}

func split(name string) (string, string) {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[1:i], name[i+1:]
	}
	return "unknown", "unknown"
}
