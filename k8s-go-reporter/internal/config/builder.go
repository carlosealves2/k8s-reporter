package config

import (
	"log"
	"os"
	"strconv"
)

// Builder implements the builder pattern for Config
type Builder struct {
	config Config
}

// NewBuilder creates a new config builder with default values
func NewBuilder() *Builder {
	return &Builder{
		config: Config{
			server: ServerConfig{
				port: "50051", // Default gRPC port
			},
		},
	}
}

// WithEnv read environment variables
func (b *Builder) WithEnv() *Builder {
	// Server configuration
	if port := os.Getenv("PORT"); port != "" {
		b.config.server.port = port
	}

	// OpenTelemetry configuration
	if getEnv("OTEL_ENABLED", false) {
		b.config.otel.enabled = true
		b.config.otel.serviceName = getEnv("OTEL_SERVICE_NAME", "k8s-go-reporter")
		b.config.otel.serviceVersion = getEnv("OTEL_SERVICE_VERSION", "unknown")
		b.config.otel.endpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
		b.config.otel.insecure = getEnv("OTEL_EXPORTER_OTLP_INSECURE", false)
		b.config.otel.tracesEnabled = getEnv("OTEL_TRACES_ENABLED", true)
		b.config.otel.metricsEnabled = getEnv("OTEL_METRICS_ENABLED", true)
	}

	return b
}

// Validate validates the configuration
// This method follows the fail-fast principle - it terminates the application if validation fails
// Returns self for method chaining
func (b *Builder) Validate() *Builder {
	// Validate port is not empty
	if b.config.server.port == "" {
		log.Fatalf("Config validation failed: PORT is required but not set")
	}

	// Validate port is numeric
	portNum, err := strconv.Atoi(b.config.server.port)
	if err != nil {
		log.Fatalf("Config validation failed: PORT '%s' must be numeric", b.config.server.port)
	}

	// Validate port is in valid range
	if portNum < 1 || portNum > 65535 {
		log.Fatalf("Config validation failed: PORT %d must be between 1 and 65535", portNum)
	}

	// Validate OpenTelemetry configuration if enabled
	if b.config.otel.enabled {
		if b.config.otel.endpoint == "" {
			log.Fatalf("Config validation failed: OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_ENABLED=true")
		}
		if !b.config.otel.tracesEnabled && !b.config.otel.metricsEnabled {
			log.Fatalf("Config validation failed: At least one of OTEL_TRACES_ENABLED or OTEL_METRICS_ENABLED must be true")
		}
	}

	return b
}

// Build returns the built configuration
func (b *Builder) Build() Config {
	return b.config
}

// Supported defines the types supported by getEnv generic function
type Supported interface {
	~string | ~int | ~bool | ~float64
}

// getEnv returns the value of an environment variable with type conversion
// If the variable is not set or conversion fails, returns the default value
func getEnv[T Supported](key string, defaultValue T) T {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	var result any

	switch any(defaultValue).(type) {
	case string:
		result = val
	case int:
		if i, err := strconv.Atoi(val); err == nil {
			result = i
		} else {
			return defaultValue
		}
	case bool:
		if b, err := strconv.ParseBool(val); err == nil {
			result = b
		} else {
			return defaultValue
		}
	case float64:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			result = f
		} else {
			return defaultValue
		}
	default:
		return defaultValue
	}

	return result.(T)
}
