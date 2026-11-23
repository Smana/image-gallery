package queue

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"

	"github.com/redis/go-redis/v9"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	obs "image-gallery/internal/observability"
)

func TestMissingQueueDoesNotFailMetricCollection(t *testing.T) {
	h := newHarness(t)
	ctx := t.Context()
	reader := sdkmetric.NewManualReader()
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("q")
	if _, err := RegisterGauges(h.rdb, meter, h.streamName(), "workers", nil); err != nil {
		t.Fatal(err)
	}
	counter, err := meter.Int64Counter("worker.other")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(ctx, 3)
	for _, scenario := range []string{"missing stream", "missing group"} {
		t.Run(scenario, func(t *testing.T) {
			if scenario == "missing group" {
				_, err := h.rdb.XAdd(ctx, &redis.XAddArgs{Stream: h.streamName(), Values: map[string]any{"job": "queued"}}).Result()
				if err != nil {
					t.Fatal(err)
				}
			}
			var rm metricdata.ResourceMetrics
			if err := reader.Collect(ctx, &rm); err != nil {
				t.Fatalf("queue failure must not abort collection: %v", err)
			}
			var other int64
			for _, scope := range rm.ScopeMetrics {
				for _, m := range scope.Metrics {
					switch m.Name {
					case obs.MetricQueueDepth, obs.MetricQueuePending, obs.MetricQueueLag:
						t.Errorf("reported %s without a working queue", m.Name)
					case "worker.other":
						sum, ok := m.Data.(metricdata.Sum[int64])
						if !ok || len(sum.DataPoints) != 1 {
							t.Fatalf("unexpected worker.other data: %#v", m.Data)
						}
						other = sum.DataPoints[0].Value
					}
				}
			}
			if other != 3 {
				t.Errorf("unrelated metric = %d, want 3", other)
			}
		})
	}
}

// TestGaugesStopPollingOnceUnregistered needs no Valkey: every poll dials, and the dialer
// only counts. After Unregister, a collection (such as the final flush on shutdown, once
// the gauge client is closed) must not poll at all.
func TestGaugesStopPollingOnceUnregistered(t *testing.T) {
	var dials atomic.Int64
	rdb := redis.NewClient(&redis.Options{MaxRetries: -1, Dialer: func(context.Context, string, string) (net.Conn, error) {
		dials.Add(1)
		return nil, errors.New("no valkey in this test")
	}})
	t.Cleanup(func() { _ = rdb.Close() })
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	reg, err := RegisterGauges(rdb, mp.Meter("q"), DefaultStream, DefaultGroup, nil)
	if err != nil {
		t.Fatal(err)
	}
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &rm); err != nil {
		t.Fatalf("unavailable Valkey must not fail collection: %v", err)
	}
	if dials.Load() == 0 {
		t.Fatal("a collection did not poll: the test cannot tell polling from not polling")
	}

	if err := reg.Unregister(); err != nil {
		t.Fatal(err)
	}
	before := dials.Load()
	if err := reader.Collect(t.Context(), &rm); err != nil {
		t.Fatal(err)
	}
	if got := dials.Load(); got != before {
		t.Errorf("a collection after Unregister still polled (%d dials)", got-before)
	}
}
