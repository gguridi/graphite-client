package graphite

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("graphite metrics", func() {

	Context("metric sum", func() {

		var (
			metric MetricSum
		)

		BeforeEach(func() {
			metric = MetricSum{}
		})

		It("should initialise with value zero", func() {
			Expect(metric.Calculate()).To(Equal("0"))
		})

		It("should update the internal value with the amount received", func() {
			metric.Update(5)
			Expect(metric.Calculate()).To(Equal("5"))
			metric.Update(3)
			Expect(metric.Calculate()).To(Equal("8"))
		})

		It("should clear the internal value", func() {
			metric.Update(5)
			metric.Clear()
			Expect(metric.Calculate()).To(Equal("0"))
		})
	})

	Context("metric average", func() {

		var (
			metric MetricAverage
		)

		BeforeEach(func() {
			metric = MetricAverage{}
		})

		It("should initialise with value zero", func() {
			Expect(metric.Calculate()).To(Equal("0"))
		})

		It("should update the internal value with the amount received", func() {
			metric.Update(2)
			Expect(metric.Calculate()).To(Equal("2.000000"))
			metric.Update(4)
			Expect(metric.Calculate()).To(Equal("3.000000"))
		})

		It("should use up to 6 decimals", func() {
			metric.Update(1)
			metric.Update(3)
			metric.Update(6)
			Expect(metric.Calculate()).To(Equal("3.333333"))
		})

		It("should clear the internal value", func() {
			metric.Update(5)
			metric.Clear()
			Expect(metric.Calculate()).To(Equal("0"))
		})
	})

	Context("metric active/inactive", func() {

		var (
			metric MetricActive
		)

		BeforeEach(func() {
			metric = MetricActive{}
		})

		It("should initialise with value inactive", func() {
			Expect(metric.Calculate()).To(Equal("0"))
		})

		It("should update the internal value with the status received", func() {
			metric.Update(true)
			Expect(metric.Calculate()).To(Equal("1"))
			metric.Update(false)
			Expect(metric.Calculate()).To(Equal("0"))
		})

		It("should clear the internal value", func() {
			metric.Update(true)
			metric.Clear()
			Expect(metric.Calculate()).To(Equal("0"))
		})
	})
})
