// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type clientReporter struct {
	monitor     *clientMonitor
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
	labels      prometheus.Labels
}

func newClientReporter(m *clientMonitor, rpcType grpcType, fullMethod string) *clientReporter {
	serviceName, methodName := splitMethodName(fullMethod)
	return &clientReporter{
		monitor:   m,
		startTime: time.Now(),
		labels: prometheus.Labels{
			labelService: serviceName,
			labelMethod:  methodName,
			labelType:    string(rpcType),
		},
	}
}

func (r *clientReporter) outgoingRequest() {
	r.monitor.requestsTotal.With(r.labels).Inc()
}

func (r *clientReporter) outgoingMessage() {
	r.monitor.messagesSent.With(r.labels).Inc()
}

func (r *clientReporter) incomingMessage() {
	r.monitor.messagesReceived.With(r.labels).Inc()
}

func (r *clientReporter) incomingResponse(code codes.Code) {
	r.labels[labelCode] = code.String()
	r.monitor.responsesTotal.With(r.labels).Inc()
	r.monitor.requestDuration.With(r.labels).Observe(time.Since(r.startTime).Seconds())
	if code != codes.OK {
		r.monitor.errors.With(r.labels).Inc()
	}
}
