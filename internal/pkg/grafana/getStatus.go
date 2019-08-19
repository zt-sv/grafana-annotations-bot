package grafana

import (
	"encoding/json"
	"github.com/go-kit/kit/log/level"
)

// HealthResp : Grafana status response
type HealthResp struct {
	Commit   string
	Database string
	Version  string
}

// GetStatus : Get Grafana status
func (client *Client) GetStatus() (HealthResp, error) {
	respJSON := HealthResp{}
	respText, err := client.apiGetRequest("/api/health", nil)

	if err != nil {
		level.Error(client.logger).Log("msg", "could not get grafana health status", "err", err)

		return respJSON, err
	}

	json.Unmarshal([]byte(respText), &respJSON)

	return respJSON, nil
}
