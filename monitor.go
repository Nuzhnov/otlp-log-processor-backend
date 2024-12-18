package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

// findAttributeInKeyValue perfroms search for an attribute key in a slice of KeyValue pairs.
func findAttributeInKeyValue(attributes []*commonpb.KeyValue, key string) (string, bool) {
	for _, attr := range attributes {
		if attr.GetKey() == key {
			// we are only interested in StringValue
			if strValue, ok := attr.Value.GetValue().(*commonpb.AnyValue_StringValue); ok {
				return strValue.StringValue, true
			}
		}
	}
	return "", false
}

type Monitor struct {
	mu       chan struct{}
	storage  map[string]int
	duration time.Duration
}

func newMonitor(duration time.Duration) *Monitor {
	return &Monitor{
		mu:       make(chan struct{}, 1),
		storage:  make(map[string]int),
		duration: duration,
	}
}

// Safely increment counter by attribute
func (m *Monitor) increment(ctx context.Context, attr string) {
	// Use buffered channel as mutex to speed up
	_, span := tracer.Start(ctx, "increment")
	defer span.End()
	span.SetAttributes(attribute.String("attributeValue", attr))
	span.AddEvent("Acquiring lock")
	m.mu <- struct{}{}
	m.storage[attr]++
	<-m.mu
	span.AddEvent("Unlocked")
}

// Print stats
func (m *Monitor) printStats(ctx context.Context) {
	lines := make([]string, 0)

	_, span := tracer.Start(ctx, "print-stats")
	defer span.End()

	span.AddEvent("Acquiring lock")
	// Read map under mutex
	m.mu <- struct{}{}
	for key := range m.storage {
		lines = append(lines, fmt.Sprintf("\"%s\" - %d", key, m.storage[key]))
	}
	<-m.mu
	span.AddEvent("Unlocked")

	if len(lines) == 0 {
		// Nothing to print
		return
	}

	// Sort  stats mutex after is free
	sort.Strings(lines)
	fmt.Println(strings.Join(lines, "\n") + "\n")
}

// Increment counter in map by value of attribute
// we seek for attribute value by "attr" attribute name on logRecord attributes,
// if it doesn't exist, we use value from Scope or Resource attributes,
// unless if it doesn't exists too, then "unknown" will be used as value
func (m *Monitor) IncrementByAttribute(ctx context.Context, req *collogspb.ExportLogsServiceRequest, attr string) {
	// Iterate over ResourceLogs
	attrName := "unknown"

	ctx, span := tracer.Start(ctx, "increment-by-attrirute")
	defer span.End()

	for _, resourceLog := range req.ResourceLogs {
		// Check Resource-level attributes
		if resource := resourceLog.GetResource(); resource != nil {
			if value, found := findAttributeInKeyValue(resource.GetAttributes(), attr); found {
				attrName = value
			}
		}

		// Iterate ScopeLogs
		for _, scopeLog := range resourceLog.GetScopeLogs() {
			// Check Scope-level attributes
			if scope := scopeLog.GetScope(); scope != nil {
				if value, found := findAttributeInKeyValue(scope.GetAttributes(), attr); found {
					attrName = value
				}
			}

			// Iterate LogRecords
			for _, logRecord := range scopeLog.GetLogRecords() {
				span.AddEvent("Incrementing attrute counter")
				// Check Log-level attributes
				if value, found := findAttributeInKeyValue(logRecord.GetAttributes(), attr); found {
					m.increment(ctx, value)
				} else {
					m.increment(ctx, attrName)
				}
			}
		}
	}
}

// Main routine for monitor - print to log stats by configured duration
func (m *Monitor) Run(ctx context.Context) {
	// Run new ticker with configured duration
	ticker := time.NewTicker(m.duration)

	ctx, span := tracer.Start(ctx, "monitoring-run")
	defer span.End()
	span.SetAttributes(attribute.String("Duration", m.duration.String()))

	slog.Info("Starting monitoring...")
	defer slog.Info("Monitoring stopped.")

	for {
		select {
		case <-ticker.C:
			m.printStats(ctx)
		case <-ctx.Done():
			return
		}
	}
}
