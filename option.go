package rc

import (
	"go.slink.ws/logging"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// region - basic client options

type BasicClientOption interface {
	Apply(*BasicClient) error
}

// region - http client

type httpClientOption struct {
	value *http.Client
}

func (q *httpClientOption) Apply(c *BasicClient) (err error) {
	c.client = q.value
	return
}

func WithHttpClient(value *http.Client) BasicClientOption {
	return &httpClientOption{
		value: value,
	}
}

// endregion - base url
// region - base url

type baseUrlOption struct {
	value string
}

func (q *baseUrlOption) Apply(c *BasicClient) (err error) {
	c.baseURL, err = url.Parse(q.value)
	return
}

func WithBaseUrl(baseUrl string) BasicClientOption {
	return &baseUrlOption{
		value: baseUrl,
	}
}

// endregion - base url
// region - user agent

type userAgentOption struct {
	value string
}

func (q *userAgentOption) Apply(c *BasicClient) error {
	c.userAgent = q.value
	return nil
}

func WithUserAgent(value string) BasicClientOption {
	return &userAgentOption{
		value: value,
	}
}

// endregion - user agent
// region - logger

type basicLoggerOption struct {
	value logging.Logger
}

func (q *basicLoggerOption) Apply(c *BasicClient) error {
	c.logger = q.value
	return nil
}

func WithBasicLogger(value logging.Logger) BasicClientOption {
	return &basicLoggerOption{
		value: value,
	}
}

// endregion - user agent

func AddMissingClientOption(opts []BasicClientOption, option BasicClientOption) []BasicClientOption {
	missingOptionType := reflect.TypeOf(option)
	for _, opt := range opts {
		existingOptType := reflect.TypeOf(opt)
		if existingOptType.String() == missingOptionType.String() {
			return opts
		}
	}
	result := make([]BasicClientOption, 0, len(opts)+1)
	for _, opt := range opts {
		result = append(result, opt)
	}
	result = append(result, option)
	return result
}

// endregion
// region - throttle client options

type ThrottleClientOption interface {
	Apply(*ThrottleClient) error
}

// region - max tokens

type maxTokensOption struct {
	value int
}

func (q *maxTokensOption) Apply(c *ThrottleClient) (err error) {
	c.maxTokens = q.value
	return
}

func WithMaxTokens(value int) ThrottleClientOption {
	return &maxTokensOption{
		value: value,
	}
}

// endregion - base url
// region - refill tokens

type refillTokensOption struct {
	value int
}

func (q *refillTokensOption) Apply(c *ThrottleClient) (err error) {
	c.maxTokens = q.value
	return
}

func WithRefillTokens(value int) ThrottleClientOption {
	return &refillTokensOption{
		value: value,
	}
}

// endregion - base url
// region - refill interval

type refillIntervalOption struct {
	value time.Duration
}

func (q *refillIntervalOption) Apply(c *ThrottleClient) (err error) {
	c.refillInterval = q.value
	return
}

func WithRefillInterval(value time.Duration) ThrottleClientOption {
	return &refillIntervalOption{
		value: value,
	}
}

// endregion - base url
// region - logger

type throttleLoggerOption struct {
	value logging.Logger
}

func (q *throttleLoggerOption) Apply(c *ThrottleClient) error {
	c.logger = q.value
	return nil
}

func WithThrottleLogger(value logging.Logger) ThrottleClientOption {
	return &throttleLoggerOption{
		value: value,
	}
}

// endregion - user agent

// endregion
// region - retry client options

type RetryClientOption interface {
	Apply(*RetryClient) error
}

// region - max attempts

type maxAttemptsOption struct {
	value int
}

func (q *maxAttemptsOption) Apply(c *RetryClient) (err error) {
	c.maxAttempts = q.value
	return
}

func WithMaxAttempts(value int) RetryClientOption {
	return &maxAttemptsOption{
		value: value,
	}
}

// endregion - base url
// region - retry delay

type retryDelayOption struct {
	value time.Duration
}

func (q *retryDelayOption) Apply(c *RetryClient) (err error) {
	c.delay = q.value
	return
}

func WithRetryDelay(value time.Duration) RetryClientOption {
	return &retryDelayOption{
		value: value,
	}
}

// endregion - base url
// region - logger

type retryLoggerOption struct {
	value logging.Logger
}

func (q *retryLoggerOption) Apply(c *RetryClient) error {
	c.logger = q.value
	return nil
}

func WithRetryLogger(value logging.Logger) RetryClientOption {
	return &retryLoggerOption{
		value: value,
	}
}

// endregion - user agent

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
// region - header

type headerOption struct {
	key   string
	value string
}

func (q *headerOption) Apply(rb *requestBuilder) error {
	if rb.headers == nil {
		rb.headers = http.Header{}
	}
	rb.headers.Set(q.key, q.value)
	return nil
}

func WithHeader(key, value string) RequestOption {
	return &headerOption{
		key:   key,
		value: value,
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
