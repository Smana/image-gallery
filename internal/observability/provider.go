package observability

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Provider manages OpenTelemetry providers lifecycle
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	config         Config
}

// NewProvider creates and initializes a new OpenTelemetry provider
func NewProvider(ctx context.Context, config Config) (*Provider, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used
		resource.WithHost(),         // Discover and provide host information
		resource.WithOS(),           // Discover and provide OS information
		resource.WithProcess(),      // Discover and provide process information
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	p := &Provider{
		config: config,
	}

	// Initialize tracer provider
	if config.TracesEnabled {
		tp, err := initTracerProvider(ctx, res, config)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
		}
		p.tracerProvider = tp
		otel.SetTracerProvider(tp)
	}

	// Initialize meter provider
	if config.MetricsEnabled {
		mp, err := initMeterProvider(ctx, res, config)
		if err != nil {
			// Clean up tracer provider if metrics fail
			if p.tracerProvider != nil {
				if shutdownErr := p.tracerProvider.Shutdown(ctx); shutdownErr != nil {
					return nil, fmt.Errorf("failed to initialize meter provider: %w (tracer shutdown also failed: %v)", err, shutdownErr)
				}
			}
			return nil, fmt.Errorf("failed to initialize meter provider: %w", err)
		}
		p.meterProvider = mp
		otel.SetMeterProvider(mp)
	}

	// Set global propagator to W3C Trace Context and Baggage
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return p, nil
}

// initTracerProvider creates and configures a tracer provider with OTLP exporter
func initTracerProvider(ctx context.Context, res *resource.Resource, config Config) (*sdktrace.TracerProvider, error) {
	// Create OTLP HTTP trace exporter
	// WithEndpointURL already specifies the scheme (http:// or https://), so WithInsecure() is not needed
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(config.TracesEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create sampler based on configuration
	sampler, err := createSampler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sampler: %w", err)
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithSampler(sampler),
	)

	return tp, nil
}

// createSampler creates a trace sampler based on configuration
func createSampler(config Config) (sdktrace.Sampler, error) {
	switch config.TracesSampler {
	case SamplerAlwaysOn:
		return sdktrace.AlwaysSample(), nil
	case SamplerAlwaysOff:
		return sdktrace.NeverSample(), nil
	case SamplerTraceIDRatio:
		ratio, err := strconv.ParseFloat(config.TracesSamplerArg, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sampler arg: %w", err)
		}
		return sdktrace.TraceIDRatioBased(ratio), nil
	case SamplerParentBasedAlwaysOn:
		return sdktrace.ParentBased(sdktrace.AlwaysSample()), nil
	case SamplerParentBasedAlwaysOff:
		return sdktrace.ParentBased(sdktrace.NeverSample()), nil
	case SamplerParentBasedTraceIDRatio:
		ratio, err := strconv.ParseFloat(config.TracesSamplerArg, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid sampler arg: %w", err)
		}
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio)), nil
	default:
		return nil, fmt.Errorf("unknown sampler type: %s", config.TracesSampler)
	}
}

// initMeterProvider creates and configures a meter provider with OTLP exporter
func initMeterProvider(ctx context.Context, res *resource.Resource, config Config) (*sdkmetric.MeterProvider, error) {
	// Create OTLP HTTP metric exporter
	// WithEndpointURL already specifies the scheme (http:// or https://), so WithInsecure() is not needed
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(config.MetricsEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create meter provider with periodic reader, exemplar filter, and views for exponential histograms
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(30*time.Second), // Export metrics every 30 seconds
		)),
		// Enable trace-based exemplar sampling - only samples measurements from sampled traces
		sdkmetric.WithExemplarFilter(exemplar.TraceBasedFilter),
		// Configure exponential histograms for all histogram metrics
		sdkmetric.WithView(createExponentialHistogramView()),
	)

	return mp, nil
}

// createExponentialHistogramView creates a view that converts all histograms to exponential histograms with exemplars
func createExponentialHistogramView() sdkmetric.View {
	return sdkmetric.NewView(
		// Match all histogram instruments (ending with .duration, .size, etc.)
		sdkmetric.Instrument{Kind: sdkmetric.InstrumentKindHistogram},
		// Convert to exponential histogram with bucket-based exemplar reservoir
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationBase2ExponentialHistogram{
				MaxSize:  160, // Maximum number of buckets (default: 160)
				MaxScale: 20,  // Maximum scale factor (default: 20, range: -10 to 20)
			},
		},
	)
}

// Tracer returns a tracer for the given instrumentation scope
func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if p.tracerProvider == nil {
		return otel.Tracer(name, opts...)
	}
	return p.tracerProvider.Tracer(name, opts...)
}

// Meter returns a meter for the given instrumentation scope
func (p *Provider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	if p.meterProvider == nil {
		return otel.Meter(name, opts...)
	}
	return p.meterProvider.Meter(name, opts...)
}

// Shutdown gracefully shuts down the provider, flushing any remaining telemetry
func (p *Provider) Shutdown(ctx context.Context) error {
	var err error

	if p.tracerProvider != nil {
		if shutdownErr := p.tracerProvider.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown tracer provider: %w", shutdownErr)
		}
	}

	if p.meterProvider != nil {
		if shutdownErr := p.meterProvider.Shutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to shutdown meter provider: %w", err, shutdownErr)
			} else {
				err = fmt.Errorf("failed to shutdown meter provider: %w", shutdownErr)
			}
		}
	}

	return err
}

// ForceFlush flushes any pending telemetry
func (p *Provider) ForceFlush(ctx context.Context) error {
	var err error

	if p.tracerProvider != nil {
		if flushErr := p.tracerProvider.ForceFlush(ctx); flushErr != nil {
			err = fmt.Errorf("failed to flush tracer provider: %w", flushErr)
		}
	}

	if p.meterProvider != nil {
		if flushErr := p.meterProvider.ForceFlush(ctx); flushErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to flush meter provider: %w", err, flushErr)
			} else {
				err = fmt.Errorf("failed to flush meter provider: %w", flushErr)
			}
		}
	}

	return err
}
