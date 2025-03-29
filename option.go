package rc

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// region - client options

type ClientOption interface {
	Apply(*Client) error
}

// region - http client

type httpClientOption struct {
	value *http.Client
}

func (q *httpClientOption) Apply(c *Client) (err error) {
	c.client = q.value
	return
}

func WithHttpClient(value *http.Client) ClientOption {
	return &httpClientOption{
		value: value,
	}
}

// endregion - base url
// region - base url

type baseUrlOption struct {
	value string
}

func (q *baseUrlOption) Apply(c *Client) (err error) {
	c.baseURL, err = url.Parse(q.value)
	return
}

func WithBaseUrl(baseUrl string) ClientOption {
	return &baseUrlOption{
		value: baseUrl,
	}
}

// endregion - base url
// region - user agent

type userAgentOption struct {
	value string
}

func (q *userAgentOption) Apply(c *Client) error {
	c.userAgent = q.value
	return nil
}

func WithUserAgent(value string) ClientOption {
	return &userAgentOption{
		value: value,
	}
}

// endregion - user agent

func AddMissingClientOption(opts []ClientOption, option ClientOption) []ClientOption {
	missingOptionType := reflect.TypeOf(option)
	for _, opt := range opts {
		existingOptType := reflect.TypeOf(opt)
		if existingOptType.String() == missingOptionType.String() {
			return opts
		}
	}
	result := make([]ClientOption, 0, len(opts)+1)
	for _, opt := range opts {
		result = append(result, opt)
	}
	result = append(result, option)
	return result
}

// endregion
// region - request options

type RequestOption interface {
	Apply(rb *requestBuilder) error
}

// region - method

type methodOption struct {
	value string
}

func (q *methodOption) Apply(rb *requestBuilder) error {
	rb.method = q.value
	return nil
}

func WithMethod(value string) RequestOption {
	return &methodOption{
		value: value,
	}
}

// endregion - method
// region - path

type pathOption struct {
	value string
}

func (q *pathOption) Apply(rb *requestBuilder) error {
	rb.queryPath = &q.value
	return nil
}

func WithQueryPath(value string) RequestOption {
	return &pathOption{
		value: value,
	}
}

// endregion - path
// region - request param

type queryParamOption struct {
	key   string
	value string
}

func (q *queryParamOption) Apply(rb *requestBuilder) error {
	if rb.queryParams == nil {
		rb.queryParams = url.Values{}
	}
	rb.queryParams.Add(q.key, q.value)
	return nil
}

func WithQueryParam(key, value string) RequestOption {
	return &queryParamOption{
		key:   key,
		value: value,
	}
}

// endregion - request param
// region - request params

type queryParamsOption struct {
	key    string
	values []string
}

func (q *queryParamsOption) Apply(rb *requestBuilder) error {
	if len(q.values) == 0 {
		return nil
	}
	if rb.queryParams == nil {
		rb.queryParams = url.Values{}
	}
	rb.queryParams.Add(q.key, strings.Join(q.values, ","))
	return nil
}

func WithQueryParamList(key string, values ...string) RequestOption {
	return &queryParamsOption{
		key:    key,
		values: values,
	}
}

// endregion - request param

// region - request body

type requestBodyOption struct {
	body any
}

func (q *requestBodyOption) Apply(rb *requestBuilder) error {
	rb.body = q.body
	return nil
}

func WithBody(body any) RequestOption {
	return &requestBodyOption{
		body: body,
	}
}

// endregion - request body

// endregion
