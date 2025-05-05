package rc

import (
	"context"
	"encoding/json"
	"go.slink.ws/logging"
	"io"
	"net/http"
	"net/url"
	"time"
)

type RetryClient struct {
	client      Client
	maxAttempts int
	delay       time.Duration
	logger      logging.Logger
}

func NewRetryClient(client Client, options ...RetryClientOption) (Client, error) {
	c := &RetryClient{
		client:      client,
		maxAttempts: -1,
		delay:       3 * time.Second,
		logger:      logging.GetNoOpLogger(),
	}
	for _, option := range options {
		option(c)
	}
	c.logger.Trace("new client")
	return c, nil
}

func (c *RetryClient) GetBaseURL() *url.URL {
	c.logger.Trace("get base url")
	return c.client.GetBaseURL()
}
func (c *RetryClient) NewRequest(options ...RequestOption) (*http.Request, error) {
	c.logger.Trace("new request")
	return c.client.NewRequest(options...)
}

func (c *RetryClient) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, int, error) {
	c.logger.Trace("do: %s %s", req.Method, req.URL)
	resp, status, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, status, err
	}
	switch v := v.(type) {
	case nil:
		return resp, status, nil
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		var b []byte
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, status, err
		}
		decErr := json.Unmarshal(b, &v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	if err != nil {
		return nil, status, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, status, err
	}
	return resp, status, err
}
func (c *RetryClient) BareDo(ctx context.Context, req *http.Request) (*Response, int, error) {

	c.logger.Trace("bare do: %s %s", req.Method, req.URL)

	if ctx == nil {
		return nil, http.StatusInternalServerError, ErrNonNilContext
	}

	var err error
	var res *Response
	var status int
	attempt := 0
	for attempt < c.maxAttempts || c.maxAttempts < 0 && ctx.Err() == nil {
		res, status, err = c.client.BareDo(ctx, req)
		if err != nil {
			switch e := err.(type) {
			case ErrTooManyRequests:
				c.logger.Debug("too many requests, wait for %v %s", e.Delay.Seconds(), "second(s)")
				time.Sleep(e.Delay)
			case ErrResourceNotFound:
				c.logger.Debug("resource not found: %s", e.Resource)
				return nil, status, err
			default:
				c.logger.Debug("error: %s, wait for %v %s", err, c.delay.Seconds(), "second(s)")
				time.Sleep(c.delay)
				attempt++
			}
			continue
		}
		return res, status, nil
	}
	return nil, status, err

}
