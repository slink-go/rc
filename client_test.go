package rc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.slink.ws/logging"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func addMissingClientOption[T any](opts []T, option T) []T {
	missingOptionType := reflect.TypeOf(option)
	for _, opt := range opts {
		existingOptType := reflect.TypeOf(opt)
		if existingOptType.String() == missingOptionType.String() {
			return opts
		}
	}
	return append(opts, option)
}
func addMissingAgent(opts []BasicClientOption) []BasicClientOption {
	return addMissingClientOption[BasicClientOption](opts, WithUserAgent("test-agent"))
}

func TestCreateClient(t *testing.T) {

	c, err := CreateClient(
		WithBasicOption(WithBaseUrl("https://google.com")),
		WithBasicOption(WithBasicLogger(logging.GetLogger("test-basic"))),
		WithThrottleOption(WithMaxTokens(10)),
		WithThrottleOption(WithRefillTokens(5)),
		WithThrottleOption(WithRefillInterval(time.Second*10)),
		WithThrottleOption(WithThrottleLogger(logging.GetLogger("test-throttle"))),
		WithRetryOption(WithMaxAttempts(-1)),
		WithRetryOption(WithRetryDelay(time.Second*10)),
		WithRetryOption(WithRetryLogger(logging.GetLogger("test-retry"))),
		WithBasicAppender(addMissingAgent),
	)
	if err != nil {
		t.Fatal(err)
	}
	rq, err := c.NewRequest(
		WithMethod("GET"),
		WithQueryPath("/"),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, st, err := c.Do(context.Background(), rq, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, st)

}
