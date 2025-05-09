package rc

import (
	"context"
	"encoding/json"
	"fmt"
	"go.slink.ws/logging"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultUserAgent = "slink http client"
)

type BasicClient struct {
	client    *http.Client
	baseURL   *url.URL
	userAgent string
	logger    logging.Logger
}

func NewBasicClient(options ...BasicClientOption) (Client, error) {

	client := &BasicClient{
		client:    http.DefaultClient,
		userAgent: defaultUserAgent,
		logger:    logging.GetNoOpLogger(),
	}

	for _, option := range options {
		option.Apply(client)
	}
	if client.baseURL == nil {
		return nil, ErrBaseUrlNotSet
	}

	client.logger.Trace("new client")

	return client, nil
}

func (c *BasicClient) GetBaseURL() *url.URL {
	c.logger.Trace("get base URL")
	if c.baseURL == nil {
		panic(ErrBaseUrlNotSet)
	}
	return c.baseURL
}

func (c *BasicClient) NewRequest(options ...RequestOption) (*http.Request, error) {

	c.logger.Trace("new request")

	rb := &requestBuilder{
		method:    http.MethodGet,
		userAgent: &c.userAgent,
		baseUrl:   c.GetBaseURL(),
	}

	for _, o := range options {
		if o != nil {
			if err := o.Apply(rb); err != nil {
				return nil, err
			}
		}
	}

	if rb.baseUrl == nil {
		return nil, ErrBaseUrlNotSet
	}

	return rb.build()

}

func (c *BasicClient) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, int, error) {

	c.logger.Debug("do: %s %s", req.Method, req.URL)

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
func (c *BasicClient) BareDo(ctx context.Context, req *http.Request) (*Response, int, error) {

	c.logger.Trace("bare do: %s %s", req.Method, req.URL)

	if ctx == nil {
		return nil, http.StatusInternalServerError, ErrNonNilContext
	}

	//fmt.Println("---------------------------------------------------------------------")
	//fmt.Println("URL: ", req.URL)
	//fmt.Println("Headers:")
	//for k, v := range req.Header {
	//	fmt.Println("  ", k, v)
	//}
	//fmt.Println("Params:")
	//for k, v := range req.URL.Query() {
	//	fmt.Println("  ", k, v)
	//}
	//fmt.Println("---------------------------------------------------------------------")

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, http.StatusRequestTimeout, ctx.Err()
		default:
		}

		// returning *url.Error.
		if e, ok := err.(*url.Error); ok {
			return nil, http.StatusRequestURITooLong, e
		}

		return nil, http.StatusBadRequest, err

	}

	response := &Response{resp}

	var status int

	err, status = checkResponse(resp)
	if err != nil {
		clErr := resp.Body.Close()
		if clErr != nil {
			return nil, status, fmt.Errorf("got some errors: \n%s \nand \n%s", err.Error(), clErr.Error())
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			// TODO: parse response to get "Retry-After"
			//       for now return -1 to make delay decision upstream
			return nil, status, ErrTooManyRequests{
				Delay: time.Duration(-1) * time.Second,
			}
		}
		return nil, status, err
	}
	return response, status, err
}
