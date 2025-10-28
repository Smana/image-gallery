# Observability

This document describes the observability features of the image-gallery application, including metrics, traces, and structured logging.

## Overview

The application uses **OpenTelemetry** to provide comprehensive observability:

- **Metrics**: Application and business metrics exported via OTLP
- **Traces**: Distributed tracing across all layers
- **Logs**: Structured JSON logs with automatic trace correlation

### Technology Stack

- **OpenTelemetry SDK**: Industry-standard observability instrumentation
- **Zerolog**: High-performance structured logging
- **OTLP Protocol**: Open standard for telemetry export
- **VictoriaMetrics**: Recommended for metrics storage (Prometheus-compatible)
- **VictoriaTraces**: Recommended for distributed tracing (Jaeger-compatible)

## Quick Start

### Local Development

1. **Start infrastructure services**:
```bash
docker-compose up -d postgres minio valkey
```

2. **Run with observability disabled** (simplest):
```bash
export OTEL_TRACES_ENABLED=false
export OTEL_METRICS_ENABLED=false
make run
```

3. **Run with observability enabled** (requires OTLP collector):
```bash
# Start a local OpenTelemetry Collector or VictoriaMetrics/VictoriaTraces
# Then run the application (it will use .env configuration)
make run
```

### Configuration

All observability features are configured via environment variables. See `.env.example` for defaults.

#### Essential Settings

```bash
# Service identification
OTEL_SERVICE_NAME=image-gallery
OTEL_SERVICE_VERSION=1.3.0
OTEL_DEPLOYMENT_ENVIRONMENT=development

# Enable/disable features
OTEL_TRACES_ENABLED=true
OTEL_METRICS_ENABLED=true

# OTLP endpoints (HTTP)
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=localhost:4318
OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=localhost:4318

# Logging
LOG_LEVEL=info        # debug, info, warn, error
LOG_FORMAT=json       # json or console
```

#### Disable Observability

To run without any observability overhead:

```bash
OTEL_TRACES_ENABLED=false
OTEL_METRICS_ENABLED=false
```

The application will start normally without connecting to any telemetry backends.

## Instrumentation Coverage

### HTTP Layer

All HTTP requests are automatically instrumented:

- **Spans**: One span per HTTP request with standard semantic conventions
- **Metrics**: Request count, duration, response size, active requests
- **Logs**: Request logs include trace_id and span_id for correlation

**Endpoints instrumented**:
- `GET /api/images` - List images with pagination
- `GET /api/images/:id` - Get specific image metadata
- `GET /api/images/:id/view` - View/download image
- `POST /api/images` - Upload new image
- `PUT /api/images/:id` - Update image metadata
- `DELETE /api/images/:id` - Delete image

### Service Layer

Business logic is instrumented with detailed traces and metrics:

- **ImageService**: Image creation, retrieval, updates, deletion
- **StorageService**: S3-compatible storage operations
- **CacheService**: Valkey cache hits/misses

### Database Layer

PostgreSQL queries are traced (when enabled in future iterations).

## Metrics Reference

### HTTP Metrics

| Metric Name | Type | Description | Labels |
|------------|------|-------------|--------|
| `http.server.request.count` | Counter | Total HTTP requests | method, route, status |
| `http.server.request.duration` | Histogram | Request duration in seconds | method, route, status |
| `http.server.response.size` | Histogram | Response size in bytes | method, route, status |
| `http.server.active_requests` | Gauge | Current active requests | - |

### Business Metrics

| Metric Name | Type | Description | Labels |
|------------|------|-------------|--------|
| `image.uploads.total` | Counter | Total image uploads | status (success/error) |
| `image.processing.duration` | Histogram | Image processing time in seconds | - |
| `cache.hits.total` | Counter | Cache hit count | operation |
| `cache.misses.total` | Counter | Cache miss count | operation |

### Storage Metrics

| Metric Name | Type | Description | Labels |
|------------|------|-------------|--------|
| `storage.operations.total` | Counter | Storage operations count | operation, status |
| `storage.operation.duration` | Histogram | Operation duration in seconds | operation |
| `storage.bytes.transferred` | Counter | Bytes uploaded/downloaded | operation |

## Traces Reference

### Trace Structure

A typical image upload trace includes:

```
HTTP Request (TracingMiddleware)
  └── listImagesHandler or createImageHandler
      └── ImageService.CreateImage
          ├── ValidationService.ValidateImage
          ├── StorageService.Store
          │   └── MinIO.PutObject
          ├── ImageProcessor.ProcessMetadata
          ├── CacheService.Set (optional)
          └── Repository.Create
              └── PostgreSQL INSERT
```

### Span Attributes

All spans include relevant business context:

- **HTTP spans**: method, route, status_code, user_agent
- **Image spans**: filename, content_type, size, width, height
- **Storage spans**: operation, bucket, object_key, size
- **Cache spans**: cache_key, hit/miss

### Span Events

Key operations emit events:

- Image validation started/completed
- Cache hit/miss events
- Processing milestones
- Error details

## Logging

### Log Format

All logs are structured JSON with consistent fields:

```json
{
  "level": "info",
  "service": "image-gallery",
  "version": "1.3.0",
  "environment": "development",
  "time": "2025-10-28T21:35:21+01:00",
  "message": "Server starting"
}
```

### Trace Correlation

When inside a trace context, logs automatically include:

```json
{
  "level": "info",
  "trace_id": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "span_id": "a1b2c3d4e5f6g7h8",
  "trace_sampled": true,
  "message": "Processing image upload"
}
```

This allows you to:
1. Find all logs for a specific trace in your log aggregation system
2. Jump from traces to logs and vice versa
3. Correlate application behavior across distributed systems

### Console vs JSON Format

**Development** (human-readable):
```bash
LOG_FORMAT=console
```

**Production** (structured):
```bash
LOG_FORMAT=json
```

## Deployment

### Kubernetes with VictoriaMetrics Operator

1. **Install VictoriaMetrics Operator**:
```bash
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm install victoria-metrics-operator vm/victoria-metrics-operator
```

2. **Deploy VictoriaMetrics components**:
```yaml
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMCluster
metadata:
  name: victoria-metrics
spec:
  retentionPeriod: "30d"
  replicationFactor: 2
  vmstorage:
    replicaCount: 2
  vmselect:
    replicaCount: 2
  vminsert:
    replicaCount: 2
```

3. **Deploy VictoriaTraces**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: victoria-traces
spec:
  template:
    spec:
      containers:
      - name: victoria-traces
        image: victoriametrics/victoria-traces:latest
        ports:
        - containerPort: 4318  # OTLP HTTP
```

4. **Configure application**:
```yaml
env:
- name: OTEL_TRACES_ENABLED
  value: "true"
- name: OTEL_METRICS_ENABLED
  value: "true"
- name: OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
  value: "victoria-traces:4318"
- name: OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
  value: "victoria-metrics-vminsert:8480"
```

### Docker Compose (Local Testing)

Add to `docker-compose.yml`:

```yaml
services:
  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4318:4318"  # OTLP HTTP
```

Example `otel-collector-config.yaml`:

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

exporters:
  logging:
    loglevel: debug
  prometheusremotewrite:
    endpoint: http://victoria-metrics:8428/api/v1/write
  jaeger:
    endpoint: victoria-traces:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging, jaeger]
    metrics:
      receivers: [otlp]
      exporters: [logging, prometheusremotewrite]
```

## Querying and Visualization

### VictoriaMetrics (PromQL)

**Request rate by endpoint**:
```promql
rate(http_server_request_count[5m])
```

**95th percentile latency**:
```promql
histogram_quantile(0.95, rate(http_server_request_duration_bucket[5m]))
```

**Image upload success rate**:
```promql
sum(rate(image_uploads_total{status="success"}[5m]))
/
sum(rate(image_uploads_total[5m]))
```

**Cache hit ratio**:
```promql
sum(rate(cache_hits_total[5m]))
/
(sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))
```

**Storage throughput**:
```promql
sum(rate(storage_bytes_transferred[5m])) by (operation)
```

### VictoriaTraces (Jaeger UI)

Access the Jaeger-compatible UI to:

1. **Find traces by service**: Filter by `image-gallery`
2. **Find slow requests**: Filter by duration > 1s
3. **Find errors**: Filter by `error=true`
4. **Trace comparison**: Compare similar operations
5. **Dependency graph**: Visualize service dependencies

### Example Queries

**Find all failed image uploads**:
- Service: `image-gallery`
- Operation: `CreateImage`
- Tags: `error=true`

**Find slow storage operations**:
- Service: `image-gallery`
- Operation: `Store`
- Min Duration: `1000ms`

## Troubleshooting

### Observability Not Working

1. **Check configuration**:
```bash
# Verify environment variables are set
env | grep OTEL
```

2. **Check logs for initialization**:
```bash
# Should see: "OpenTelemetry provider initialized"
tail -f /tmp/server.log
```

3. **Check OTLP endpoint connectivity**:
```bash
# Test if endpoint is reachable
curl -v http://localhost:4318/v1/traces
```

### High Overhead

If observability is causing performance issues:

1. **Reduce trace sampling**:
```bash
# Sample 10% of traces (configure in future iteration)
OTEL_TRACES_SAMPLER=traceidratio
OTEL_TRACES_SAMPLER_ARG=0.1
```

2. **Disable metrics**:
```bash
OTEL_METRICS_ENABLED=false
```

3. **Adjust log level**:
```bash
LOG_LEVEL=warn  # Only log warnings and errors
```

### Connection Refused Errors

If you see:
```
failed to upload metrics: Post "http://localhost:4318/": connect: connection refused
```

This is **expected** when the OTLP endpoint is not available. The application uses graceful degradation and will continue running normally. To fix:

1. Start an OTLP collector (see Deployment section)
2. Or disable observability if not needed

## Best Practices

### Development

- Use `LOG_FORMAT=console` for readable logs
- Use `LOG_LEVEL=debug` for detailed information
- Enable observability only when actively debugging
- Use local OTLP collector with logging exporter

### Production

- Use `LOG_FORMAT=json` for structured logging
- Use `LOG_LEVEL=info` or `LOG_LEVEL=warn`
- Always enable observability in production
- Configure appropriate retention policies
- Set up alerting on key metrics
- Monitor trace sampling rates

### Monitoring

Key metrics to alert on:

- `http_server_request_count{status=~"5.."}` - Server errors
- `image_uploads_total{status="error"}` - Failed uploads
- `http_server_request_duration` p95 > 1s - Slow requests
- `storage_operations_total{status="error"}` - Storage failures
- `cache_hits_total / (cache_hits_total + cache_misses_total)` < 0.5 - Low cache hit ratio

## Architecture

### Observability Components

```
┌─────────────────────────────────────────────────────────┐
│                   Application Code                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ HTTP Handler │  │    Service   │  │  Repository  │  │
│  │  (Traced)    │→ │   (Traced)   │→ │   (Traced)   │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│         ↓                  ↓                  ↓          │
│  ┌────────────────────────────────────────────────────┐ │
│  │         OpenTelemetry SDK (Provider)               │ │
│  │  • TracerProvider  • MeterProvider  • Logger       │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                          ↓
                 ┌────────────────┐
                 │  OTLP Exporter │
                 │   (HTTP/4318)  │
                 └────────────────┘
                          ↓
         ┌────────────────┴────────────────┐
         ↓                                  ↓
┌──────────────────┐            ┌──────────────────┐
│ VictoriaMetrics  │            │ VictoriaTraces   │
│  (Metrics Store) │            │  (Trace Store)   │
└──────────────────┘            └──────────────────┘
```

### Graceful Degradation

The application is designed to function normally even when observability backends are unavailable:

1. **Provider initialization fails**: Application continues without tracing/metrics
2. **OTLP endpoint unreachable**: Telemetry is buffered then dropped
3. **Export failures**: Logged but don't affect business logic
4. **Shutdown**: ForceFlush ensures buffered telemetry is exported

## Further Reading

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [VictoriaMetrics Documentation](https://docs.victoriametrics.com/)
- [VictoriaTraces Documentation](https://docs.victoriametrics.com/victorialogs/)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
- [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
