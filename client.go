package rc

import (
	"context"
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
