package bhaserver

import (
	"bin-vul-inspector/pkg/client"
)

type Client struct {
	*client.Client
}

func NewClient(baseUrl string, options ...client.Option) *Client {
	return &Client{
		Client: client.NewClient(baseUrl, options...),
	}
}

func (c *Client) Scan(params ScanReq) (id string, err error) {
	var body ScanResp

	u, err := c.FullUrl("/bha/scan")
	if err != nil {
		return "", err
	}

	res, err := c.HttpClient.R().SetBody(params).SetError(&body).SetResult(&body).Post(u)
	if err != nil {
		return "", err
	}
	if err = c.CheckResponse(res, body.Response); err != nil {
		return "", err
	}
	return body.Data.Id, nil
}

func (c *Client) Terminate(id string) (err error) {
	var body client.Response

	u, err := c.FullUrl("/bha/stop")
	if err != nil {
		return err
	}

	res, err := c.HttpClient.R().
		SetPathParams(map[string]string{"id": id}).
		SetError(&body).
		SetResult(&body).
		Post(u)
	if err != nil {
		return err
	}
	if err = c.CheckResponse(res, body); err != nil {
		return err
	}
	return nil
}
