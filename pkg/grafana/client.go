package grafana

import (
	"crypto/tls"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// Client : Grafana client
type Client struct {
	grafanaURL *url.URL
	token      string
	useTLS     bool
	skipVerify bool
	caFile     string
	certFile   string
	keyFile    string
	logger     log.Logger
}

// ClientConfig : Grafana client config
type ClientConfig struct {
	URL        *url.URL
	Token      string
	UseTLS     bool
	SkipVerify bool
	CertFile   string
	KeyFile    string
	Logger     log.Logger
}

// NewClient : create new Grafana client
func NewClient(config ClientConfig) (*Client, error) {
	client := &Client{
		grafanaURL: config.URL,
		token:      config.Token,
		useTLS:     config.UseTLS,
		skipVerify: config.SkipVerify,
		certFile:   config.CertFile,
		keyFile:    config.KeyFile,
		logger:     config.Logger,
	}

	return client, nil
}

func (client *Client) getHTTPClient() *http.Client {
	if client.useTLS {
		cert, err := tls.LoadX509KeyPair(client.certFile, client.keyFile)

		if err != nil {
			level.Error(client.logger).Log("msg", "failed to load keys", "err", err)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: client.skipVerify,
			},
		}

		return &http.Client{Transport: tr}
	}

	return &http.Client{}
}

func (client *Client) getEndpointURL(endpoint string, query map[string]string) string {
	uri, _ := url.Parse(client.grafanaURL.String())
	uri.Path = path.Join(uri.Path, endpoint)

	q := uri.Query()

	for k := range query {
		q.Set(k, query[k])
	}

	uri.RawQuery = q.Encode()

	return uri.String()
}

func (client *Client) apiGetRequest(apiPath string, query map[string]string) (string, error) {
	var (
		endpoint = client.getEndpointURL(apiPath, query)
	)

	httpClient := client.getHTTPClient()
	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("Authorization", "Bearer "+client.token)
	resp, err := httpClient.Do(req)

	if err != nil {
		level.Error(client.logger).Log("msg", "could not get request to "+endpoint, "err", err)

		return string(""), err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			level.Error(client.logger).Log("msg", "error parsing response body", "err", err)
		}

		return string(bodyBytes), nil
	}

	level.Error(client.logger).Log("msg", "request to "+endpoint+" finished with error ", "err", err, "status code", resp.StatusCode)
	return string(""), err
}

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
