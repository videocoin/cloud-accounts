package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	url      string
	user     string
	password string
	cli      *http.Client
}

func NewClient(u string, opts ...Option) *Client {
	c := &Client{
		url: u,
		cli: &http.Client{},
	}

	urlParsed, err := url.Parse(u)
	if err == nil {
		if urlParsed.User.Username() != "" {
			if passwd, ok := urlParsed.User.Password(); ok {
				c.user = urlParsed.User.Username()
				c.password = passwd
				urlParsed.User = nil
				c.url = urlParsed.String()
			}
		}
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *Client) Do(account string, amount uint64) error {
	data, err := json.Marshal(map[string]interface{}{"account": account, "amount": amount})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	if c.user != "" && c.password != "" {
		req.SetBasicAuth(c.user, c.password)
	}

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error %d", resp.StatusCode)
	}

	return nil
}
