package grafana

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Client : Grafana client
type Client struct {
	grafanaURL  *url.URL
	token       string
	TLSInsecure bool
	skipVerify  bool
	caFile      string
	certFile    string
	keyFile     string
	logger      log.Logger
}

// ClientConfig : Grafana client config
type ClientConfig struct {
	URL         *url.URL
	Token       string
	TLSInsecure bool
	SkipVerify  bool
	CertFile    string
	KeyFile     string
	Logger      log.Logger
}

// NewClient : create new Grafana client
func NewClient(config ClientConfig) (*Client, error) {
	client := &Client{
		grafanaURL:  config.URL,
		token:       config.Token,
		TLSInsecure: config.TLSInsecure,
		skipVerify:  config.SkipVerify,
		certFile:    config.CertFile,
		keyFile:     config.KeyFile,
		logger:      config.Logger,
	}

	return client, nil
}

func (client *Client) getHTTPClient() *http.Client {
	if !client.TLSInsecure {
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
