package rc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.slink.ws/logging"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ErrResourceNotFound struct {
	Resource string
}

func (e ErrResourceNotFound) Error() string {
	return fmt.Sprintf("Resource Not Found: %s", e.Resource)
}

type ErrTooManyRequests struct {
	Delay time.Duration
}

func (e ErrTooManyRequests) Error() string {
	return fmt.Sprintf("Too Many Requests; Wait %s", e.Delay)
}

type callerFunc func(ctx context.Context, req *http.Request) (*Response, int, error)

type ThrottleClient struct {
	client         Client
	maxTokens      int
	refillTokens   int
	refillInterval time.Duration
	logger         logging.Logger
	caller         callerFunc
}

func NewThrottleClient(client Client, options ...ThrottleClientOption) (Client, error) {

	c := &ThrottleClient{
		client:         client,
		maxTokens:      30,
		refillTokens:   30,
		refillInterval: time.Minute,
		logger:         logging.GetNoOpLogger(),
	}

	for _, option := range options {
		option(c)
	}

	c.caller = c.throttler(c.client.BareDo)
	c.logger.Trace("new client")

	return c, nil
}

func (c *ThrottleClient) GetBaseURL() *url.URL {
	c.logger.Trace("get base url")
	return c.client.GetBaseURL()
}
func (c *ThrottleClient) NewRequest(options ...RequestOption) (*http.Request, error) {
	c.logger.Trace("new request")
	return c.client.NewRequest(options...)
}

func (c *ThrottleClient) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, int, error) {
	c.logger.Trace("do: %s %s", req.Method, req.URL)
	resp, status, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, status, err
	}
	switch v := v.(type) {
	case nil:
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

func (c *ThrottleClient) BareDo(ctx context.Context, req *http.Request) (*Response, int, error) {
	c.logger.Trace("bare do: %s %s", req.Method, req.URL)
	//return c.throttler(c.client.BareDo)(ctx, req)
	return c.caller(ctx, req)
}

func (c *ThrottleClient) throttler(e callerFunc) callerFunc {

	c.logger.Trace("throttler enter")
	defer c.logger.Trace("throttler exit")

	var tokens = c.maxTokens
	var once sync.Once
	var lastRefill time.Time

	return func(ctx context.Context, req *http.Request) (*Response, int, error) {
		c.logger.Trace("throttler closure: enter")
		defer c.logger.Trace("throttler closure: exit")

		if ctx == nil {
			return nil, http.StatusInternalServerError, ErrNonNilContext
		}

		if ctx.Err() != nil {
			return nil, http.StatusRequestTimeout, ctx.Err()
		}
		once.Do(func() {
			lastRefill = time.Now()
			ticker := time.NewTicker(c.refillInterval)
			go func() {
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						lastRefill = time.Now()
						if tokens < c.maxTokens {
							c.logger.Debug("throttler: refill tokens")
							t := tokens + c.refillTokens
							if t > c.maxTokens {
								t = c.maxTokens
							}
							tokens = t
						}
					}
				}
			}()
		})
		if tokens <= 0 {
			delay := c.calculateDelay(lastRefill)
			c.logger.Debug("throttler: wait for %v %s", delay.Seconds(), "second(s)")
			return nil, http.StatusTooManyRequests, ErrTooManyRequests{
				Delay: delay,
			}
		}
		c.logger.Trace("throttler: call external service")
		tokens--

		res, status, err := e(ctx, req)

		// if our throttling was not enough, and we received 429 error from external service
		if errors.As(err, &ErrTooManyRequests{}) {
			c.logger.Warning("throttler: too many requests error from external service")
			tokens /= 5 // чтобы не в ноль сбрасывать; чтобы по возможности ждать не весь refillInterval
			return nil, http.StatusTooManyRequests, ErrTooManyRequests{
				Delay: c.calculateDelay(lastRefill) / 2,
			}
		}

		return res, status, err
	}
}
func (c *ThrottleClient) calculateDelay(lastRefill time.Time) time.Duration {
	ms := c.refillInterval.Milliseconds() - time.Now().Sub(lastRefill).Milliseconds() + 100
	if ms < 0 {
		ms = 0
	}
	return time.Duration(ms) * time.Millisecond
}
