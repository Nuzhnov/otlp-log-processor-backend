package main

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type dash0LogsServiceServer struct {
	addr      string
	attribute string
	monitor   *Monitor

	collogspb.UnimplementedLogsServiceServer
}

func newServer(addr, attribute string, monitor *Monitor) collogspb.LogsServiceServer {
	s := &dash0LogsServiceServer{
		addr:      addr,
		attribute: attribute,
		monitor:   monitor,
	}
	return s
}

func (l *dash0LogsServiceServer) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	slog.DebugContext(ctx, "Received ExportLogsServiceRequest")
	logsReceivedCounter.Add(ctx, 1)

	ctx, span := tracer.Start(ctx, "handle-export-log-service-request")
	defer span.End()

	span.SetAttributes(attribute.String("attributeName", l.attribute))

	l.monitor.IncrementByAttribute(ctx, request, l.attribute)

	return &collogspb.ExportLogsServiceResponse{}, nil
}
