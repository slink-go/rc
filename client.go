package rc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	defaultUserAgent = "slink http client"
)

type Client struct {
	client    *http.Client
	baseURL   *url.URL
	userAgent string
}

func NewClient(options ...ClientOption) (*Client, error) {

	client := &Client{
		client:    http.DefaultClient,
		userAgent: defaultUserAgent,
	}

	for _, option := range options {
		if option != nil {
			if err := option.Apply(client); err != nil {
				return nil, err
			}
		}
	}
	if client.baseURL == nil {
		return nil, ErrBaseUrlNotSet
	}

	return client, nil
}

func (c *Client) GetBaseURL() *url.URL {
	if c.baseURL == nil {
		panic(ErrBaseUrlNotSet)
	}
	return c.baseURL
}

func (c *Client) NewRequest(options ...RequestOption) (*http.Request, error) {

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

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		var b []byte
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
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
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return resp, err
}
func (c *Client) BareDo(ctx context.Context, req *http.Request) (*Response, error) {
	if ctx == nil {
		return nil, ErrNonNilContext
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
			return nil, ctx.Err()
		default:
		}

		// returning *url.Error.
		if e, ok := err.(*url.Error); ok {
			return nil, e
		}

		return nil, err

	}

	response := &Response{resp}

	err = c.checkResponse(resp)
	if err != nil {
		clErr := resp.Body.Close()
		if clErr != nil {
			return nil, fmt.Errorf("got some errors: \n%s \nand \n%s", err.Error(), clErr.Error())
		}
		return nil, err
	}
	return response, err
}
func (c *Client) checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	return fmt.Errorf("status:[%d] %s", r.StatusCode, r.Status)
}
