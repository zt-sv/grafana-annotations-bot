package grafana

import (
	"encoding/json"
	"github.com/go-kit/kit/log/level"
	"strconv"
	"time"
)

// Annotation : grafana annotation
type Annotation struct {
	ID          int
	AlertID     int
	DashboardID int
	PanelID     int
	UserID      int
	UserName    string
	NewState    string
	PrevState   string
	Time        int64
	Text        string
	Metric      string
	RegionID    int
	Tags        []string
}

// AnnotationsResp : Grafana annotations list
type AnnotationsResp []Annotation

// GetAnnotations : get annotations list from Grafana
func (client *Client) GetAnnotations(fromTime time.Time, toTime time.Time) (AnnotationsResp, error) {
	respJSON := AnnotationsResp{}

	query := map[string]string{}
	query["from"] = strconv.FormatInt(fromTime.UnixNano()/int64(time.Millisecond), 10)
	query["to"] = strconv.FormatInt(toTime.UnixNano()/int64(time.Millisecond), 10)

	respText, err := client.apiGetRequest("/api/annotations", query)

	if err != nil {
		level.Error(client.logger).Log("msg", "could not get grafana annotations", "err", err)
		return respJSON, err
	}

	json.Unmarshal([]byte(respText), &respJSON)

	return respJSON, nil
}
