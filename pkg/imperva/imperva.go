package imperva

import (
	"github.com/buger/jsonparser"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"

	//"io"
	"net/http"
	"net/url"
	"time"
)

const minBefore = 5

const bandwidthTotalQuery = "api/stats/v1"

var (
	ImpervaURL    string
	ImpervaApiId  string
	ImpervaApiKey string
	ImpervaSiteId string
)

type myTransport struct{}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("x-API-Id", ImpervaApiId)
	req.Header.Add("x-API-Key", ImpervaApiKey)
	return http.DefaultTransport.RoundTrip(req)
}

// CreateClient
func CreateClient(impervaURL, impervaApiId, impervaApiKey, impervaSiteId string) (*http.Client, error) {

	ImpervaURL = impervaURL
	ImpervaApiId = impervaApiId
	ImpervaApiKey = impervaApiKey
	ImpervaSiteId = impervaSiteId
	httpClient := http.Client{
		Transport: &myTransport{},
		Timeout:   time.Second * 10, // Timeout after 2 seconds
	}
	return &httpClient, nil
}

func QueryBandwidthTotal(client *http.Client) (float64, error) {

	data := url.Values{
		"site_id":     {ImpervaSiteId},
		"stats":       {"bandwidth_timeseries"},
		"time_range":  {"custom"},
		"start":       {strconv.FormatInt(1000*time.Now().Add(-1*time.Hour).Unix(), 10)},
		"end":         {strconv.FormatInt(1000*time.Now().Add(0*time.Hour).Unix(), 10)},
		"granularity": {"300000"},
	}

	resp, err := client.PostForm(ImpervaURL, data)

	if err != nil {
		log.Fatal(err)
	}

	jsonByte, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var maxTime, metricTime, bandwidth int64
	maxTime = 0

	jsonparser.ArrayEach(jsonByte, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		key, _ := jsonparser.GetString(value, "id")
		switch key {
		case "api.stats.bandwidth_timeseries.bandwidth":
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				metricTime, _ = jsonparser.GetInt(value, "[0]")
				//	fmt.Println(metricTime)
				if metricTime > maxTime {
					bandwidth, _ = jsonparser.GetInt(value, "[1]")
					//		fmt.Println(metricTime, maxTime, bandwidth)
					maxTime = metricTime
				}
			}, "data")
		}
	}, "bandwidth_timeseries")
	log.Info(bandwidth, time.UnixMilli(maxTime).UTC())
	return float64(bandwidth), nil
}
