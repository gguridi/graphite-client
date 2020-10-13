package graphite

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("graphite aggregator", func() {

	var (
		client     Graphite
		agg        Aggregator
		testMetric = "metric"
		failMetric = "force.fail"
		mutex      = &sync.Mutex{}
	)

	var getTotalSent = func(client Graphite) int {
		mutex.Lock()
		defer mutex.Unlock()
		total := 0
		for _, value := range client.(*MockGraphite).Data {
			number, _ := strconv.Atoi(value)
			total += number
		}
		return total
	}

	var getFlushSent = func(client Graphite) int {
		mutex.Lock()
		defer mutex.Unlock()
		return len(client.(*MockGraphite).Data)
	}

	var getMetricInfo = func(received string) (string, int, int) {
		var metric string
		var value int
		var timestamp int
		fmt.Sscanf(received, "%s %d %d\n", &metric, &value, &timestamp)
		return metric, value, timestamp
	}

	BeforeEach(func() {
		client = &MockGraphite{
			Extra: map[string]interface{}{},
			Data:  map[string]string{},
			MethodReconnect: func(m *MockGraphite) error {
				m.Extra["reconnected"] = "yes, it tried!"
				return nil
			},
			MethodSendBuffer: func(m *MockGraphite, buffer *bytes.Buffer) (int, error) {
				mutex.Lock()
				defer mutex.Unlock()
				metric, value, _ := getMetricInfo(buffer.String())
				entry := strconv.FormatInt(time.Now().UnixNano(), 10)
				m.Data[entry] = strconv.FormatInt(int64(value), 10)
				if metric == failMetric {
					return 0, errors.New("Unable to send metrics to graphite")
				}
				return 0, nil
			},
		}
		agg = &aggregator{
			config:  &Config{},
			client:  client,
			metrics: map[string]Metric{},
		}
	})

	It("should implement Aggregator interface", func() {
		var _ Aggregator = (*aggregator)(nil)
	})

	It("is returned by the graphite client properly configured", func() {
		client = NewGraphiteTCP(&Config{Host: "test.com"})
		aggregator := client.NewAggregator().(*aggregator)
		Expect(aggregator.config).To(Equal(client.(*graphite).config))
		Expect(aggregator.client).To(Equal(client))
	})

	Context("sum aggregates", func() {

		It("should initialise metric with first given value", func() {
			agg.AddSum(testMetric, 5)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("5"))
		})

		It("should sum subsequent metrics", func() {
			agg.AddSum(testMetric, 5)
			agg.AddSum(testMetric, 3)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("8"))
		})

		It("is thread-safe", func() {
			var wg sync.WaitGroup
			wg.Add(200)
			for i := 0; i < 200; i++ {
				go func(value int) {
					agg.AddSum(testMetric, value)
					wg.Done()
				}(i)
			}
			wg.Wait()
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("19900"))
		})
	})

	Context("increase aggregates", func() {

		It("should initialise metric with 1", func() {
			agg.Increase(testMetric)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("1"))
		})

		It("should increase subsequent calls by 1", func() {
			for i := 0; i < 200; i++ {
				agg.Increase(testMetric)
			}
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("200"))
		})

		It("is thread-safe", func() {
			var wg sync.WaitGroup
			wg.Add(200)
			for i := 0; i < 200; i++ {
				go func() {
					agg.Increase(testMetric)
					wg.Done()
				}()
			}
			wg.Wait()
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("200"))
		})
	})

	Context("average aggregates", func() {

		It("should initialise metric with first given value", func() {
			agg.AddAverage(testMetric, 5)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("5.000000"))
		})

		It("should calculate the average from the subsequent metrics", func() {
			agg.AddAverage(testMetric, 5)
			agg.AddAverage(testMetric, 3)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("4.000000"))
		})

		It("is thread-safe", func() {
			var wg sync.WaitGroup
			wg.Add(200)
			for i := 0; i < 200; i++ {
				go func(value int) {
					agg.AddAverage(testMetric, value)
					wg.Done()
				}(i)
			}
			wg.Wait()
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("99.500000"))
		})
	})

	Context("active/inactive aggregates", func() {

		It("should set metric to active", func() {
			agg.SetActive(testMetric)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("1"))
		})

		It("should set metric to inactive", func() {
			agg.SetInactive(testMetric)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("0"))
		})

		It("should keep the latest metric set", func() {
			agg.SetActive(testMetric)
			agg.SetInactive(testMetric)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(Equal("0"))
		})

		It("is thread-safe", func() {
			var wg sync.WaitGroup
			wg.Add(200)
			for i := 0; i < 200; i++ {
				go func(value int) {
					if value%2 == 0 {
						agg.SetActive(testMetric)
					} else {
						agg.SetInactive(testMetric)
					}
					wg.Done()
				}(i)
			}
			wg.Wait()
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[testMetric].Calculate()).To(SatisfyAny(Equal("0"), Equal("1")))
		})
	})

	Context("flushes the aggregates to send them to graphite", func() {

		It("is thread-safe", func() {
			var wg sync.WaitGroup
			wg.Add(200)
			for i := 0; i < 200; i++ {
				go func(value int) {
					agg.AddSum(testMetric, value)
					if value%50 == 0 {
						agg.Flush()
					}
					wg.Done()
				}(i)
			}
			wg.Wait()
			agg.Flush()
			Expect(getTotalSent(client)).To(Equal(19900))
		})
	})

	Context("runs periodically", func() {

		var (
			stop chan bool
		)

		BeforeEach(func() {
			stop = make(chan bool)
			agg = &aggregator{
				config:  &Config{},
				client:  client,
				metrics: map[string]Metric{},
			}
		})

		It("flushes every tick", func() {
			agg.Run(2*time.Second, stop)
			agg.AddSum(testMetric, 15)
			agg.AddSum(testMetric, 25)
			time.Sleep(3 * time.Second)
			Expect(getTotalSent(client)).To(Equal(40))
			agg.AddSum(testMetric, 30)
			time.Sleep(3 * time.Second)
			Expect(getTotalSent(client)).To(Equal(70))
			stop <- true
		})

		It("uses a chan bool to stop the periodic flushing", func() {
			agg.Run(2*time.Second, stop)
			agg.AddSum(testMetric, 15)
			time.Sleep(3 * time.Second)
			Expect(getTotalSent(client)).To(Equal(15))
			stop <- true
			agg.AddSum(testMetric, 25)
			time.Sleep(3 * time.Second)
			Expect(getTotalSent(client)).To(Equal(15))
		})

		It("keeps trying periodically even if one Flush fails, keeping the data", func() {
			agg.Run(2*time.Second, stop)
			agg.AddSum(failMetric, 15)
			time.Sleep(3 * time.Second)
			Expect(getFlushSent(client)).To(Equal(2))
			agg.AddSum(failMetric, 25)
			time.Sleep(2 * time.Second)
			Expect(getFlushSent(client)).To(Equal(4))
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics[failMetric].Calculate()).To(Equal("40"))
		})

		It("retries once if something went wrong", func() {
			agg.Run(2*time.Second, stop)
			agg.AddSum(failMetric, 15)
			time.Sleep(3 * time.Second)
			Expect(getFlushSent(client)).To(Equal(2))
			Expect(client.(*MockGraphite).Extra).To(HaveKey("reconnected"))
		})
	})

	Context("uses client configuration", func() {

		It("uses the namespace/prefix before sending metrics", func() {
			agg = &aggregator{
				config: &Config{
					Namespace: "beta.instance",
				},
				client:  client,
				metrics: map[string]Metric{},
			}
			agg.AddSum(testMetric, 5000)
			metrics := agg.(*aggregator).GetMetrics()
			Expect(metrics["beta.instance."+testMetric].Calculate()).To(Equal("5000"))
		})
	})
})
