package riot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	circuitBreakerAPIEndpoint = "/v1/api/health"
)

// RFC-100 HealthEnum describes the overall health of a service.
type HealthEnum string

const (
	HealthEnabled  HealthEnum = "ENABLED"
	HealthDegraded HealthEnum = "DEGRADED"
	HealthDisabled HealthEnum = "DISABLED"
)

// HealthStatus is the DTO returned from a query/status request which contains the overall
// HealthEnum status, a reason string, and optional details string to string map.
type HealthStatus struct {
	Status  HealthEnum        `json:"status"`
	Reason  string            `json:"reason"`
	Details map[string]string `json:"details"`
}

func IsRTPHealthy() (bool, error) {
	circuitBreakerURL := os.Getenv("RTP_CIRCUIT_BREAKER_ENDPOINT")

	resp, err := http.Get(fmt.Sprintf("%s/%s", circuitBreakerURL, circuitBreakerAPIEndpoint))
	if err != nil {
		return true, fmt.Errorf("Failed to get RTP Circuit Breaker Health, assuming true for health.")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, fmt.Errorf("Failed to get RTP Circuit Breaker Response Body, assuming true for health.")
	}

	rtpCircuitBreakerStatus := HealthStatus{}
	err = json.Unmarshal(body, &rtpCircuitBreakerStatus)
	if err != nil {
		return true, fmt.Errorf("Failed to parse RTP Circuit Breaker Response, assuming true for health.")
	}

	return (rtpCircuitBreakerStatus.Status == HealthEnabled), nil
}
