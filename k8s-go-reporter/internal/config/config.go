package config

// Config holds all application configuration
type Config struct {
	server ServerConfig
	otel   OTelConfig
}

// ServerConfig contains server-specific configuration
type ServerConfig struct {
	port string
}

// OTelConfig contains OpenTelemetry configuration
type OTelConfig struct {
	enabled        bool
	serviceName    string
	serviceVersion string
	endpoint       string
	insecure       bool
	tracesEnabled  bool
	metricsEnabled bool
}

// Server returns the server configuration
func (c *Config) Server() ServerConfig {
	return c.server
}

// OTel returns the OpenTelemetry configuration
func (c *Config) OTel() OTelConfig {
	return c.otel
}

// Port returns the server port
func (s ServerConfig) Port() string {
	return s.port
}

// Enabled returns whether OpenTelemetry is enabled
func (o OTelConfig) Enabled() bool {
	return o.enabled
}

// ServiceName returns the service name for OpenTelemetry
func (o OTelConfig) ServiceName() string {
	return o.serviceName
}

// ServiceVersion returns the service version for OpenTelemetry
func (o OTelConfig) ServiceVersion() string {
	return o.serviceVersion
}

// Endpoint returns the OTLP exporter endpoint
func (o OTelConfig) Endpoint() string {
	return o.endpoint
}

// Insecure returns whether to use insecure connection
func (o OTelConfig) Insecure() bool {
	return o.insecure
}

// TracesEnabled returns whether traces are enabled
func (o OTelConfig) TracesEnabled() bool {
	return o.tracesEnabled
}

// MetricsEnabled returns whether metrics are enabled
func (o OTelConfig) MetricsEnabled() bool {
	return o.metricsEnabled
}
