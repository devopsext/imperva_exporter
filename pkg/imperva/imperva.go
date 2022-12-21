package imperva

import (
	"fmt"
	"github.com/buger/jsonparser"
	sreCommon "github.com/devopsext/sre/common"
	"io/ioutil"
	"strconv"

	//"io"
	"net/http"
	"net/url"
	"time"
)

const minBefore = 5

const bandwidthQuery = "/api/stats/v1"
const siteListQuery = "/api/prov/v1/sites/list"

var (
	ImpervaFQDN   string
	ImpervaApiId  string
	ImpervaApiKey string
	ImpervaSiteId string
	Logger        sreCommon.Logger
)

type ApiStatsBandwidthTimeseries struct {
	BandwidthValue                             float64
	BandwidthTime                              time.Time
	Value__api_stats_bandwidth_timeseries__bps float64
	BPSTime                                    time.Time
}

type ImpervaClient struct {
	httpClient *http.Client
	logger     sreCommon.Logger
}

type myTransport struct{}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("x-API-Id", ImpervaApiId)
	req.Header.Add("x-API-Key", ImpervaApiKey)
	return http.DefaultTransport.RoundTrip(req)
}

// CreateClient
func CreateClient(impervaFQDN, impervaApiId, impervaApiKey string, logger sreCommon.Logger) (*http.Client, error) {
	Logger = logger
	ImpervaFQDN = impervaFQDN
	ImpervaApiId = impervaApiId
	ImpervaApiKey = impervaApiKey
	httpClient := http.Client{
		Transport: &myTransport{},
		Timeout:   time.Second * 10, // Timeout after 2 seconds
	}
	return &httpClient, nil
}

func QuerySiteIdMap(client *http.Client) (map[int64]string, error) {
	Logger.Debug("QuerySiteIdMap")
	siteList := make(map[int64]string)
	Logger.Debug("post form")
	resp, err := client.PostForm("https://"+ImpervaFQDN+siteListQuery, url.Values{})
	if err != nil {
		Logger.Panic(err)
	}
	jsonByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	Logger.Debug("parse result")
	jsonparser.ArrayEach(jsonByte, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		site_id, _ := jsonparser.GetInt(value, "site_id")
		domain, _ := jsonparser.GetString(value, "domain")
		siteList[site_id] = domain
	}, "sites")
	Logger.Debug(fmt.Sprintf("%v", siteList))
	return siteList, nil
}

func Query__bandwidth_timeseries(client *http.Client, impervaSiteId string) (ApiStatsBandwidthTimeseries, error) {
	Logger.Debug(fmt.Sprintf("QueryBandwidth for site_id=%s", impervaSiteId))
	data := url.Values{
		"site_id":     {impervaSiteId},
		"stats":       {"bandwidth_timeseries"},
		"time_range":  {"custom"},
		"start":       {strconv.FormatInt(1000*time.Now().Add(-1*time.Hour).Unix(), 10)},
		"end":         {strconv.FormatInt(1000*time.Now().Add(0*time.Hour).Unix(), 10)},
		"granularity": {"300000"},
	}

	resp, err := client.PostForm("https://"+ImpervaFQDN+bandwidthQuery, data)
	if err != nil {
		Logger.Panic(err)
	}

	jsonByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var maxTimeBandwidth, maxTimeBPS, metricTime, bandwidth, Value__api_stats_bandwidth_timeseries__bps int64
	maxTimeBandwidth = 0
	maxTimeBPS = 0

	jsonparser.ArrayEach(jsonByte, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		key, _ := jsonparser.GetString(value, "id")
		switch key {
		case "api.stats.bandwidth_timeseries.bandwidth":
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				metricTime, _ = jsonparser.GetInt(value, "[0]")
				//	fmt.Println(metricTime)
				if metricTime > maxTimeBandwidth {
					bandwidth, _ = jsonparser.GetInt(value, "[1]")
					//		fmt.Println(metricTime, maxTime, bandwidth)
					maxTimeBandwidth = metricTime
				}
			}, "data")
		case "api.stats.bandwidth_timeseries.bps":
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				metricTime, _ = jsonparser.GetInt(value, "[0]")
				//	fmt.Println(metricTime)
				if metricTime > maxTimeBPS {
					Value__api_stats_bandwidth_timeseries__bps, _ = jsonparser.GetInt(value, "[1]")
					//		fmt.Println(metricTime, maxTime, bandwidth)
					maxTimeBPS = metricTime
				}
			}, "data")
		}
	}, "bandwidth_timeseries")
	return ApiStatsBandwidthTimeseries{
		BandwidthValue: float64(bandwidth),
		BandwidthTime:  time.UnixMilli(maxTimeBandwidth),
		Value__api_stats_bandwidth_timeseries__bps: float64(Value__api_stats_bandwidth_timeseries__bps),
		BPSTime: time.UnixMilli(maxTimeBPS),
	}, nil
}
