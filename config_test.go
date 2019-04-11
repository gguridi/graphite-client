package graphite

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("configuration", func() {

	var (
		config Config
	)

	Context("metrics path", func() {

		var (
			testPrefix = "alpha.instance"
		)

		BeforeEach(func() {
			config = Config{
				Namespace: testPrefix,
			}
		})

		It("appends the prefix if set in the configuration", func() {
			metricPath := "additional-path"
			Expect(config.getMetricPath(metricPath)).To(Equal(testPrefix + "." + metricPath))
		})

		It("uses the prefix directly if no metric name is passed", func() {
			Expect(config.getMetricPath("")).To(Equal(testPrefix))
		})

		It("uses the metric name directly if no prefix is configured", func() {
			metricPath := "metric.path.count"
			config.Namespace = ""
			Expect(config.getMetricPath(metricPath)).To(Equal(metricPath))
		})
	})

	Context("graphite address", func() {

		BeforeEach(func() {
			config = Config{
				Host: "example.com",
				Port: 2003,
			}
		})

		It("constructs properly the endpoint string given a host and a port", func() {
			Expect(config.getAddress()).To(Equal("example.com:2003"))
		})

		It("doesn't trigger an error if host is not set", func() {
			config.Host = ""
			Expect(config.getAddress()).To(Equal(":2003"))
		})

		It("doesn't trigger an error if port is not set", func() {
			config.Port = 0
			Expect(config.getAddress()).To(Equal("example.com:0"))
		})
	})

	Context("client timeout", func() {

		BeforeEach(func() {
			config = Config{}
		})

		It("returns a default timeout of one seconds if none is provided", func() {
			Expect(config.getTimeout()).To(Equal(1 * time.Second))
		})

		It("returns the timeout set as time duration", func() {
			config.Timeout = 5 * time.Minute
			Expect(config.getTimeout()).To(Equal(config.Timeout))
		})
	})
})
