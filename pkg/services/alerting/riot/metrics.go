package riot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Metric struct {
	Name     string `json:"name"`
	Hostname string `json:"rfc460Hostname"`
	Instance string `json:"instance"`

	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"rfc460Timestamp"`

	Scope string `json:"rfc190Scope"`
}

// ConstructMetric generates a new Metric
func ConstructMetric(name string, value interface{}) Metric {
	hostname := os.Getenv("RC_HOSTNAME")

	scope := fmt.Sprintf("%s.%s.%s", os.Getenv("RC_ENVIRONMENT"), os.Getenv("RC_DATACENTER"), os.Getenv("RC_SHARD"))

	return Metric{
		Name:      name,
		Value:     value,
		Hostname:  hostname,
		Instance:  "grafana",
		Scope:     scope,
		Timestamp: time.Now(),
	}
}

// Send sends a metric to RTP
func (m Metric) Send() error {
	metrics := []Metric{m}
	payloadBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	metricURL := os.Getenv("RTP_COLLECTOR_METRIC_ENDPOINT")
	if metricURL != "" {
		req, err := http.NewRequest("POST", metricURL, body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	return nil
}
