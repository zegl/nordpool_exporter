package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PriceResponse struct {
	Data []struct {
		ID    string  `json:"id"`
		Value float64 `json:"value"`
	} `json:"Data"`
	LastUpdated float64 `json:"LastUpdated"`
}

type nordPoolCollector struct {
	logger             *zap.Logger
	price              *prometheus.GaugeVec
	priceScrapesFailed prometheus.Counter
}

var variableGroupLabelNames = []string{
	"id",
}

func NewNordpoolCollector(namespace string, logger *zap.Logger) prometheus.Collector {
	c := &nordPoolCollector{
		logger: logger,
		price: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "price",
				Name:      "price",
				Help:      "Prices",
			},
			variableGroupLabelNames,
		),
		priceScrapesFailed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "price",
				Name:      "scrapes_failed",
				Help:      "Count of scrapes of group data from NordPool that have failed",
			},
		),
	}

	return c
}

func (c nordPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	c.price.Describe(ch)
}

func (c *nordPoolCollector) Collect(ch chan<- prometheus.Metric) {
	c.price.Reset()

	if prices, err := c.getPrices(); err != nil {
		c.logger.Error("Failed to update prices", zap.Error(err))
		c.priceScrapesFailed.Inc()
	} else {
		for _, price := range prices.Data {
			c.price.With(prometheus.Labels{"id": price.ID}).Set(price.Value)
		}
	}

	c.price.Collect(ch)
}

func (c *nordPoolCollector) getPrices() (*PriceResponse, error) {
	url := fmt.Sprintf("https://www.svk.se/services/controlroom/v2/map/price?ticks=%d", time.Now().UnixMilli())
	log.Printf("Fetching %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read prices: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("%s", body)

	var res PriceResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prices: %w", err)
	}

	return &res, nil
}
