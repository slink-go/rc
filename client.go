package rc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// region - client option

// endregion
// region - client interface

type Client interface {
	GetBaseURL() *url.URL
	NewRequest(options ...RequestOption) (*http.Request, error)
	Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error)
	BareDo(ctx context.Context, req *http.Request) (*Response, error)
}

// endregion
// region - client util

func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	return fmt.Errorf("status:[%d] %s", r.StatusCode, r.Status)
}

// endregion
