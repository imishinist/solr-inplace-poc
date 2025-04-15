package solr

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

type Client struct {
	host       string
	collection string

	httpClient *http.Client
}

func NewClient(host string, collection string) *Client {
	return &Client{
		host:       host,
		collection: collection,
		httpClient: http.DefaultClient,
	}
}

func (c *Client) url(component string, params url.Values) string {
	u := path.Join("solr", c.collection, component)
	return fmt.Sprintf("http://%s/%s?%s", c.host, u, params.Encode())
}

func (c *Client) Update(body string) (io.ReadCloser, error) {
	params := url.Values{}
	params.Add("commit", "true")
	params.Add("failOnVersionConflicts", "false")

	url := c.url("update", params)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
