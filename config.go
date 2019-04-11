package graphite

import (
	"fmt"
	"time"
)

const (
	// DefaultTimeout specifies the default timeout to use when connecting to graphite.
	DefaultTimeout = 1 * time.Second
)

// Config stores the configuration to pass to the graphite client.
type Config struct {
	// Host is a string specifying the address where graphite is listening. It can be
	// a host name "example.com" or an IP address "219.123.43.21". This field is required.
	Host string
	// Port is an integer specifying the port where graphite is listening. This field is required.
	Port int
	// Namespace specifies a prefix to use for all the metrics, so we don't need to set it
	// every time we want to send something.
	Namespace string
	// Timeout specifies a new timeout in time.Duration format in case we want to increase/decrease
	// the default one. Defaults to 1 second.
	Timeout time.Duration
	// ForceReconnect is a boolean specifying if we want to force a reconnection every time we send metrics
	// to graphite. This is useful when working with AWS ELB or any other network components that might
	// be tampering with the connections.
	ForceReconnect bool
}

func (config *Config) getMetricPath(metricPath string) string {
	if config.Namespace != "" && metricPath != "" {
		return fmt.Sprintf("%s.%s", config.Namespace, metricPath)
	}
	return config.Namespace + metricPath
}

func (config *Config) getAddress() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}

func (config *Config) getTimeout() time.Duration {
	if config.Timeout > 0 {
		return config.Timeout
	}
	return DefaultTimeout
}
