package main

import (
	"context"
	"testing"
	"time"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	respb "go.opentelemetry.io/proto/otlp/resource/v1"
)

// Helper function to create a mock ExportLogsServiceRequest
func createMockRequest() *collogspb.ExportLogsServiceRequest {
	return &collogspb.ExportLogsServiceRequest{
		ResourceLogs: []*logspb.ResourceLogs{
			{
				Resource: &respb.Resource{
					Attributes: []*commonpb.KeyValue{
						{Key: "env", Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "production"}}},
					},
				},
				ScopeLogs: []*logspb.ScopeLogs{
					{
						Scope: &commonpb.InstrumentationScope{
							Attributes: []*commonpb.KeyValue{
								{Key: "version", Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "v1.0"}}},
							},
						},
						LogRecords: []*logspb.LogRecord{
							{
								Attributes: []*commonpb.KeyValue{
									{Key: "service", Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "user-service"}}},
								},
							},
							{
								Attributes: []*commonpb.KeyValue{
									{Key: "service", Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: "order-service"}}},
								},
							},
							{
								Attributes: []*commonpb.KeyValue{},
							},
						},
					},
				},
			},
		},
	}
}

func TestIncrementByLogsAttribute(t *testing.T) {
	monitor := newMonitor(1 * time.Second)
	req := createMockRequest()

	ctx := context.Background()
	monitor.IncrementByAttribute(ctx, req, "service")

	if len(monitor.storage) != 3 {
		t.Fatalf("Expected 3 unique attributes, got %d", len(monitor.storage))
	}

	if monitor.storage["user-service"] != 1 {
		t.Errorf("Expected \"user-service\" count to be 1, got %d", monitor.storage["user-service"])
	}

	if monitor.storage["order-service"] != 1 {
		t.Errorf("Expected \"order-service\" count to be 1, got %d", monitor.storage["order-service"])
	}

	if monitor.storage["unknown"] != 1 {
		t.Errorf("Expected \"unknown\" count to be 1, got %d", monitor.storage["unknown"])
	}
}

func TestIncrementByScopeAttribute(t *testing.T) {
	monitor := newMonitor(1 * time.Second)
	req := createMockRequest()

	ctx := context.Background()
	monitor.IncrementByAttribute(ctx, req, "version")

	if len(monitor.storage) != 1 {
		t.Fatalf("Expected 1 unique attribute, got %d", len(monitor.storage))
	}

	if monitor.storage["v1.0"] != 3 {
		t.Errorf("Expected \"v1.0\" count to be 1, got %d", monitor.storage["v1.0"])
	}
}

func TestIncrementByResourceAttribute(t *testing.T) {
	monitor := newMonitor(1 * time.Second)
	req := createMockRequest()

	ctx := context.Background()
	monitor.IncrementByAttribute(ctx, req, "env")

	if len(monitor.storage) != 1 {
		t.Fatalf("Expected 1 unique attribute, got %d", len(monitor.storage))
	}

	if monitor.storage["production"] != 3 {
		t.Errorf("Expected \"production\" count to be 1, got %d", monitor.storage["production"])
	}
}

func TestIncrementByWrongAttribute(t *testing.T) {
	monitor := newMonitor(1 * time.Second)
	req := createMockRequest()

	ctx := context.Background()
	monitor.IncrementByAttribute(ctx, req, "doesnotexist")

	if len(monitor.storage) != 1 {
		t.Fatalf("Expected 1 unique attribute, got %d", len(monitor.storage))
	}

	if monitor.storage["unknown"] != 3 {
		t.Errorf("Expected \"unknown\" count to be 1, got %d", monitor.storage["unknown"])
	}
}
