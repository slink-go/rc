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
	Apply(*BasicClient)
}

type basicOptionHttpClient struct {
	value *http.Client
}

func (o *basicOptionHttpClient) Apply(client *BasicClient) {
	client.client = o.value
}

func WithHttpClient(value *http.Client) BasicClientOption {
	return &basicOptionHttpClient{value}
}

type basicOptionBaseUrl struct {
	value string
}

func (o *basicOptionBaseUrl) Apply(client *BasicClient) {
	value, err := url.Parse(o.value)
	if err != nil {
		panic(err)
	}
	client.baseURL = value
}

func WithBaseUrl(baseUrl string) BasicClientOption {
	return &basicOptionBaseUrl{baseUrl}
}

type basicUserAgent struct {
	value string
}

func (o *basicUserAgent) Apply(client *BasicClient) {
	client.userAgent = o.value
}

func WithUserAgent(value string) BasicClientOption {
	return &basicUserAgent{value}
}

type basicLogger struct {
	value logging.Logger
}

func (o *basicLogger) Apply(client *BasicClient) {
	client.logger = o.value
}
func WithBasicLogger(value logging.Logger) BasicClientOption {
	return &basicLogger{value}
}

func AddMissingBasicClientOption(opts []BasicClientOption, option BasicClientOption) []BasicClientOption {
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

type ThrottleClientOption func(*ThrottleClient)

func WithMaxTokens(value int) ThrottleClientOption {
	return func(client *ThrottleClient) {
		client.maxTokens = value
	}
}
func WithRefillTokens(value int) ThrottleClientOption {
	return func(client *ThrottleClient) {
		client.refillTokens = value
	}
}
func WithRefillInterval(value time.Duration) ThrottleClientOption {
	return func(client *ThrottleClient) {
		client.refillInterval = value
	}
}
func WithThrottleLogger(value logging.Logger) ThrottleClientOption {
	return func(client *ThrottleClient) {
		client.logger = value
	}
}

// endregion
// region - retry client options

type RetryClientOption func(*RetryClient)

func WithMaxAttempts(value int) RetryClientOption {
	return func(client *RetryClient) {
		client.maxAttempts = value
	}
}
func WithRetryDelay(value time.Duration) RetryClientOption {
	return func(client *RetryClient) {
		client.delay = value
	}
}
func WithRetryLogger(value logging.Logger) RetryClientOption {
	return func(client *RetryClient) {
		client.logger = value
	}
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
