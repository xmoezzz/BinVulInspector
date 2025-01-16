package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	baseUrl    string
	HttpClient *resty.Client
}

type Option func(*Client)

func WithInsecureSkipVerify(insecureTlsCert bool) Option {
	return func(c *Client) {
		c.HttpClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: insecureTlsCert})
	}
}

func WithToken(token string) Option {
	return func(c *Client) {
		c.HttpClient.SetAuthToken(token)
	}
}

func NewClient(baseUrl string, options ...Option) (c *Client) {
	c = &Client{
		baseUrl:    baseUrl,
		HttpClient: resty.New(),
	}
	for _, option := range options {
		option(c)
	}

	return c
}

func (c *Client) FullUrl(format string, a ...any) (result string, err error) {
	return url.JoinPath(c.baseUrl, fmt.Sprintf(format, a...))
}

func (c *Client) CheckResponse(res *resty.Response, body Response) error {
	if res.StatusCode() == http.StatusOK && body.Code == StatusOk {
		return nil
	}

	if body.ErrMessage != "" {
		return errors.New(body.ErrMessage)
	}
	if res.Status() != "" {
		return errors.New(res.Status())
	}
	return errors.New(http.StatusText(res.StatusCode()))
}
