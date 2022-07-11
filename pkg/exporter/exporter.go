package exporter

import (
	"errors"
	"imperva_exporter/pkg/imperva"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const namespace = "imperva"

type ImpervaExporter struct {
	ListenAddress string
	MetricsPath   string
	ImpervaURL    string
	ImpervaApiId  string
	ImpervaApiKey string
	ImpervaSiteId string
}

var (

	// Metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last Imperva query successful.",
		nil, nil,
	)
	bandwidthTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "bandwidth_total"),
		"Total amount of traffic for SiteID.",
		nil,
		nil,
	)
)

func NewExporter(options ImpervaExporter) (*ImpervaExporter, error) {
	if options.ImpervaApiId == "" || options.ImpervaApiKey == "" {
		return nil, errors.New("no credentials provided")
	}
	return &ImpervaExporter{
		ListenAddress: options.ListenAddress,
		MetricsPath:   options.MetricsPath,
		ImpervaURL:    options.ImpervaURL,
		ImpervaApiId:  options.ImpervaApiId,
		ImpervaApiKey: options.ImpervaApiKey,
		ImpervaSiteId: options.ImpervaSiteId,
	}, nil
}

func (e *ImpervaExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- bandwidthTotal
}

func (e *ImpervaExporter) Collect(ch chan<- prometheus.Metric) {
	log.Debug("Start scraping...")

	impervaClient, err := imperva.CreateClient(
		e.ImpervaURL,
		e.ImpervaApiId,
		e.ImpervaApiKey,
		e.ImpervaSiteId,
	)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Error(err)
		return
	}

	curBandwidthTotal, err := imperva.QueryBandwidthTotal(impervaClient)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Error(err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		bandwidthTotal, prometheus.GaugeValue, curBandwidthTotal,
	)

	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)

	log.Debug("Scraping has successfully finished")
}
