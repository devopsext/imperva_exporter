package exporter

import (
	"container/list"
	"errors"
	"fmt"
	sreCommon "github.com/devopsext/sre/common"
	"github.com/prometheus/client_golang/prometheus"
	"imperva_exporter/pkg/imperva"
	"strconv"
	"sync"
	"time"
)

const namespace = "imperva"

type Entity struct {
	domain  string
	site_id string
	stat    imperva.ApiStatsBandwidthTimeseries
}

var Entities = []Entity{}
var entities = []Entity{}

type ImpervaExporter struct {
	ListenAddress string
	MetricsPath   string
	ImpervaFQDN   string
	ImpervaApiId  string
	ImpervaApiKey string
	MuScrap       sync.Mutex
	MuCopy        sync.Mutex
	logger        sreCommon.Logger
	//SiteIDs       	[]string{"domain", "site_id"}
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
		[]string{"domain", "site_id"},
		nil,
	)
	api_stats_bandwidth_timeseries__bps = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "api_stats_bandwidth_timeseries__bps"),
		"BPS traffic for SiteID.",
		[]string{"domain", "site_id"},
		nil,
	)
)

func NewExporter(options ImpervaExporter, logger *sreCommon.Logs) (*ImpervaExporter, error) {
	l := list.New()
	fmt.Println(l)

	logger.Debug("NewExporter")
	if options.ImpervaApiId == "" || options.ImpervaApiKey == "" {
		return nil, errors.New("no credentials provided")
	}
	return &ImpervaExporter{
		ListenAddress: options.ListenAddress,
		MetricsPath:   options.MetricsPath,
		ImpervaFQDN:   options.ImpervaFQDN,
		ImpervaApiId:  options.ImpervaApiId,
		ImpervaApiKey: options.ImpervaApiKey,
		logger:        logger,
	}, nil
}

func (e *ImpervaExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- bandwidthTotal
	ch <- api_stats_bandwidth_timeseries__bps
}

func (e *ImpervaExporter) LoopCollect() {
	entities = nil
	e.logger.Debug("Start scraping...")

	impervaClient, _ := imperva.CreateClient(
		e.ImpervaFQDN,
		e.ImpervaApiId,
		e.ImpervaApiKey,
		e.logger,
	)

	QuerySiteIdMap, _ := imperva.QuerySiteIdMap(impervaClient)

	for id, domain := range QuerySiteIdMap {
		//go func() {
		stat, _ := imperva.Query__bandwidth_timeseries(impervaClient, strconv.FormatInt(id, 10))

		entities = append(entities, Entity{
			domain:  domain,
			site_id: strconv.FormatInt(id, 10),
			stat:    stat,
		})
	}
	e.logger.Debug("MuCopy.Lock before copy into Entities")
	e.MuCopy.Lock()
	defer e.MuCopy.Unlock()
	Entities = nil
	for _, v := range entities {
		Entities = append(Entities, v)
	}
	e.logger.Debug("MuCopy.Unlock after copy into Entities")
}

func (e *ImpervaExporter) Collect(ch chan<- prometheus.Metric) {
	//entities = nil
	//e.LoopCollect()
	//
	//
	//e.logger.Debug("Start scraping...")
	//
	//impervaClient, err := imperva.CreateClient(
	//	e.ImpervaFQDN,
	//	e.ImpervaApiId,
	//	e.ImpervaApiKey,
	//	e.logger,
	//)
	//if err != nil {
	//	ch <- prometheus.MustNewConstMetric(
	//		up, prometheus.GaugeValue, 0,
	//	)
	//	e.logger.Error(err)
	//	return
	//}
	//
	//QuerySiteIdMap, err := imperva.QuerySiteIdMap(impervaClient)
	//if err != nil {
	//	ch <- prometheus.MustNewConstMetric(
	//		up, prometheus.GaugeValue, 0,
	//	)
	//	e.logger.Error(err)
	//	return
	//}

	//for id, domain := range QuerySiteIdMap {
	//	//go func() {
	//	timestamp, curBandwidthTotal, err := imperva.QueryBandwidthTotal(impervaClient, strconv.FormatInt(id, 10))
	//	if err != nil {
	//		ch <- prometheus.MustNewConstMetric(
	//			up, prometheus.GaugeValue, 0,
	//		)
	//		e.logger.Error(err)
	//		return
	//	}
	//	if timestamp != time.UnixMilli(0) {
	//		ch <- prometheus.NewMetricWithTimestamp(timestamp,
	//			prometheus.MustNewConstMetric(
	//				bandwidthTotal, prometheus.GaugeValue, curBandwidthTotal, domain, strconv.FormatInt(id, 10),
	//			))
	//	}
	//	//}()
	//}
	e.logger.Debug("MuCopy.Lock before metric scraping")
	e.MuCopy.Lock()

	defer e.MuCopy.Unlock()
	for _, v := range Entities {
		if v.stat.BandwidthTime != time.UnixMilli(0) {
			ch <- prometheus.MustNewConstMetric(
				bandwidthTotal, prometheus.GaugeValue, v.stat.BandwidthValue, v.domain, v.site_id,
			)
			ch <- prometheus.MustNewConstMetric(
				api_stats_bandwidth_timeseries__bps, prometheus.GaugeValue, v.stat.Value__api_stats_bandwidth_timeseries__bps, v.domain, v.site_id,
			)
			//ch <- prometheus.NewMetricWithTimestamp(v.timestamp,
			//	prometheus.MustNewConstMetric(
			//		bandwidthTotal, prometheus.GaugeValue, v.value, v.domain, v.site_id,
			//	))
		}

	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)

	e.logger.Debug("MuCopy.unLock after metrics scraping ")
}
