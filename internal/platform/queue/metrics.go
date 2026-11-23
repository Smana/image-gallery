package queue

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	obs "image-gallery/internal/observability"
)

// RegisterGauges reports the consumer group's backlog on each collection.
// Pass an UNINSTRUMENTED client: these polls must not show up as traces.
//
//	queue.depth   entries not yet delivered to the group (XINFO GROUPS lag; XLEN also counts acked entries)
//	queue.pending delivered but not acked
//	queue.lag     age of the oldest pending entry, in seconds
//
// Unregister the returned registration before closing rdb, or the final
// collection on shutdown polls a closed client.
// Failed polls skip the gauges and are logged when log is non-nil.
func RegisterGauges(rdb redis.UniversalClient, meter metric.Meter, stream, group string, log *obs.Logger) (metric.Registration, error) {
	depth, err := meter.Int64ObservableGauge(obs.MetricQueueDepth, metric.WithUnit("{message}"),
		metric.WithDescription("Entries not yet delivered to the consumer group"))
	if err != nil {
		return nil, err
	}
	pending, err := meter.Int64ObservableGauge(obs.MetricQueuePending, metric.WithUnit("{message}"),
		metric.WithDescription("Entries delivered but not acknowledged"))
	if err != nil {
		return nil, err
	}
	lag, err := meter.Float64ObservableGauge(obs.MetricQueueLag, metric.WithUnit("s"),
		metric.WithDescription("Age of the oldest pending entry"))
	if err != nil {
		return nil, err
	}
	attrs := metric.WithAttributes(semconv.MessagingDestinationName(stream), semconv.MessagingConsumerGroupName(group))
	return meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		backlog, err := readBacklog(ctx, rdb, stream, group)
		if err != nil {
			if log != nil && ctx.Err() != context.Canceled {
				log.Warn(ctx).Err(err).Str("stream", stream).Str("group", group).Msg("queue metrics unavailable")
			}
			// A callback error prevents the periodic reader from exporting any metrics.
			return nil
		}
		if backlog.depth >= 0 {
			o.ObserveInt64(depth, backlog.depth, attrs)
		}
		o.ObserveInt64(pending, backlog.pending, attrs)
		o.ObserveFloat64(lag, backlog.age, attrs)
		return nil
	}, depth, pending, lag)
}

type backlog struct {
	depth   int64
	pending int64
	age     float64
}

func readBacklog(ctx context.Context, rdb redis.UniversalClient, stream, group string) (backlog, error) {
	groups, err := rdb.XInfoGroups(ctx, stream).Result()
	if err != nil {
		return backlog{}, err
	}
	sum, err := rdb.XPending(ctx, stream, group).Result()
	if err != nil {
		return backlog{}, err
	}
	b := backlog{depth: -1, pending: sum.Count}
	for _, g := range groups {
		if g.Name == group {
			b.depth = g.Lag
			break
		}
	}
	if sum.Count > 0 {
		b.age = entryAge(sum.Lower, time.Now())
	}
	return b, nil
}
