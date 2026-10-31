package rc

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type stubClient struct{}

func (stubClient) GetBaseURL() *url.URL { return nil }

func (stubClient) NewRequest(options ...RequestOption) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, "http://stub/", nil)
}

func (stubClient) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, int, error) {
	return stubClient{}.BareDo(ctx, req)
}

func (stubClient) BareDo(ctx context.Context, req *http.Request) (*Response, int, error) {
	return &Response{&http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}}, http.StatusOK, nil
}

// TestThrottleClient_RefillOutlivesRequestContext verifies that the throttler
// refill goroutine keeps running after the context of the first request is
// canceled. Previously the refill goroutine was bound to that context, so once
// the first request completed the token bucket froze at zero forever.
func TestThrottleClient_RefillOutlivesRequestContext(t *testing.T) {
	t.Parallel()

	throttle, err := NewThrottleClient(
		stubClient{},
		WithMaxTokens(1),
		WithRefillTokens(1),
		WithRefillInterval(50*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	c, err := NewRetryClient(throttle)
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.NewRequest()
	if err != nil {
		t.Fatal(err)
	}

	ctx1, cancel1 := context.WithCancel(context.Background())
	if _, _, err := c.Do(ctx1, req, nil); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	cancel1()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	done := make(chan error, 1)
	go func() {
		_, _, err := c.Do(ctx2, req, nil)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("second request failed after refill: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second request did not succeed: refill goroutine died with the first request's context")
	}
}

func TestThrottleClient_CalculateDelayFloor(t *testing.T) {
	t.Parallel()

	c := &ThrottleClient{refillInterval: time.Second}
	if got := c.calculateDelay(time.Now().Add(-time.Hour)); got < 100*time.Millisecond {
		t.Fatalf("expected delay >= 100ms, got %v", got)
	}
}

type throttledOnceClient struct{ calls int }

func (c *throttledOnceClient) GetBaseURL() *url.URL { return nil }
func (c *throttledOnceClient) NewRequest(options ...RequestOption) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, "http://stub/", nil)
}
func (c *throttledOnceClient) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, int, error) {
	return c.BareDo(ctx, req)
}
func (c *throttledOnceClient) BareDo(ctx context.Context, req *http.Request) (*Response, int, error) {
	c.calls++
	if c.calls == 1 {
		return nil, http.StatusTooManyRequests, ErrTooManyRequests{Delay: 0}
	}
	return &Response{&http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Request: req}}, http.StatusOK, nil
}

// TestRetryClient_ZeroTooManyRequestsDelayFallsBack verifies that a
// non-positive ErrTooManyRequests delay does not degrade into an immediate
// busy-loop retry.
func TestRetryClient_ZeroTooManyRequestsDelayFallsBack(t *testing.T) {
	t.Parallel()

	inner := &throttledOnceClient{}
	c, err := NewRetryClient(inner, WithRetryDelay(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	req, err := c.NewRequest()
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	if _, _, err := c.Do(context.Background(), req, nil); err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Fatalf("expected fallback delay >= 100ms, got %v", elapsed)
	}
	if inner.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", inner.calls)
	}
}
