package riot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
)

var (
	circuitBreakerAPIEndpoint = "/v1/api/healthByCluster"
)

// RFC-100 HealthEnum describes the overall health of a service.
type HealthEnum string

const (
	HealthEnabled                HealthEnum = "ENABLED"
	HealthDegraded               HealthEnum = "DEGRADED"
	HealthDisabled               HealthEnum = "DISABLED"
	RTP_CIRCUIT_BREAKER_ENDPOINT            = "RTP_CIRCUIT_BREAKER_ENDPOINT"
)

// HealthStatus is the DTO returned from a query/status request which contains the overall
// HealthEnum status, a reason string, and optional details string to string map.
type HealthStatus struct {
	Status  HealthEnum        `json:"status"`
	Reason  string            `json:"reason"`
	Details map[string]string `json:"details"`
}

type HealthByCluster map[string]*HealthStatus

func IsRTPHealthy(dataSourceUrl string) (bool, error) {
	circuitBreakerURL := os.Getenv(RTP_CIRCUIT_BREAKER_ENDPOINT)
	url, err := url.Parse(dataSourceUrl)
	if err != nil {
		return true, fmt.Errorf("Failed to process RTP Circuit Breaker Health, assuming true for health. Cannot parse datasource url %s", dataSourceUrl)
	}

	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}

	resp, err := http.Get(fmt.Sprintf("%s/%s", circuitBreakerURL, circuitBreakerAPIEndpoint))
	if err != nil {
		return true, fmt.Errorf("Failed to get RTP Circuit Breaker Health, assuming true for health.")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, fmt.Errorf("Failed to get RTP Circuit Breaker Response Body, assuming true for health.")
	}
	rtpHealthByCluster := HealthByCluster{}
	err = json.Unmarshal(body, &rtpHealthByCluster)
	if err != nil {
		return true, fmt.Errorf("Failed to parse RTP Circuit Breaker Response, assuming true for health.")
	}
	return rtpHealthByCluster[host] == nil || rtpHealthByCluster[host].Status != HealthDegraded, nil
}
