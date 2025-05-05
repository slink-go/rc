package rc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// region - client option

// endregion
// region - client interface

type Client interface {
	GetBaseURL() *url.URL
	NewRequest(options ...RequestOption) (*http.Request, error)
	Do(ctx context.Context, req *http.Request, v interface{}) (*Response, int, error)
	BareDo(ctx context.Context, req *http.Request) (*Response, int, error)
}

// endregion
// region - client util

func checkResponse(r *http.Response) (error, int) {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil, r.StatusCode
	}

	if r.StatusCode == http.StatusTooManyRequests {
		// TODO: parse response to get "Retry-After"
		//       for now return -1 to make delay decision upstream
		return ErrTooManyRequests{
			Delay: time.Duration(-1) * time.Second,
		}, http.StatusTooManyRequests
	}

	if r.StatusCode == http.StatusNotFound {
		return ErrResourceNotFound{
			r.Request.URL.String(),
		}, http.StatusNotFound
	}

	return fmt.Errorf("status:[%d] %s", r.StatusCode, r.Status), r.StatusCode
}

// endregion
// region - client builder

// region - rest client config

type rcConfig struct {
	basicOptions     []BasicClientOption
	throttleOptions  []ThrottleClientOption
	retryOptions     []RetryClientOption
	basicAppender    []func(options []BasicClientOption) []BasicClientOption
	throttleAppender []func(options []ThrottleClientOption) []ThrottleClientOption
	retryAppender    []func(options []RetryClientOption) []RetryClientOption
}

// endregion
// region - option

type RestClientOption func(*rcConfig)

// directly provide client options

func WithBasicOption(option BasicClientOption) RestClientOption {
	return func(config *rcConfig) {
		config.basicOptions = append(config.basicOptions, option)
	}
}
func WithThrottleOption(option ThrottleClientOption) RestClientOption {
	return func(config *rcConfig) {
		config.throttleOptions = append(config.throttleOptions, option)
	}
}
func WithRetryOption(option RetryClientOption) RestClientOption {
	return func(config *rcConfig) {
		config.retryOptions = append(config.retryOptions, option)
	}
}

// provide 'client options provider'

func WithBasicAppender(appender func(options []BasicClientOption) []BasicClientOption) RestClientOption {
	return func(config *rcConfig) {
		config.basicAppender = append(config.basicAppender, appender)
	}
}
func WithThrottleAppender(appender func(options []ThrottleClientOption) []ThrottleClientOption) RestClientOption {
	return func(config *rcConfig) {
		config.throttleAppender = append(config.throttleAppender)
	}
}
func WithRetryAppender(appender func(options []RetryClientOption) []RetryClientOption) RestClientOption {
	return func(config *rcConfig) {
		config.retryAppender = append(config.retryAppender, appender)
	}
}

// endregion

func CreateClient(options ...RestClientOption) (Client, error) {

	cfg := &rcConfig{}

	for _, option := range options {
		option(cfg)
	}
	if len(cfg.basicOptions) == 0 && len(cfg.basicAppender) == 0 {
		return nil, errors.New("basic rest client options not provided")
	}

	for _, a := range cfg.basicAppender {
		cfg.basicOptions = a(cfg.basicOptions)
	}

	client, err := NewBasicClient(cfg.basicOptions...)
	if err != nil {
		return nil, err
	}

	if len(cfg.throttleOptions) > 0 || len(cfg.throttleOptions) > 0 {
		for _, a := range cfg.throttleAppender {
			cfg.throttleOptions = a(cfg.throttleOptions)
		}
		client, err = NewThrottleClient(client, cfg.throttleOptions...)
		if err != nil {
			return nil, err
		}
	}
	if len(cfg.retryOptions) > 0 || len(cfg.retryAppender) > 0 {
		for _, a := range cfg.retryAppender {
			cfg.retryOptions = a(cfg.retryOptions)
		}
		client, err = NewRetryClient(client, cfg.retryOptions...)
		if err != nil {
			return nil, err
		}
	}

	return client, err

}

// endregion
