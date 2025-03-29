package rc

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func createTestClient() *Client {
	c, err := NewClient(
		WithBaseUrl("https://test.com"),
		WithUserAgent("test-agent"),
	)
	if err != nil {
		panic(err)
	}
	return c
}

func TestCreteRequestBasic(t *testing.T) {
	req, err := createTestClient().NewRequest()
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com", req.URL.String())
}
func TestCreteRequestWithPath(t *testing.T) {
	req, err := createTestClient().NewRequest(WithQueryPath("/test"))
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/test", req.URL.String())
}
func TestCreteRequestWithMethod(t *testing.T) {
	req, err := createTestClient().NewRequest(
		WithMethod(http.MethodHead),
		WithQueryPath("/test"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/test", req.URL.String())
	assert.Equal(t, http.MethodHead, req.Method)
}
func TestCreteRequestWithUserAgent(t *testing.T) {
	req, err := createTestClient().NewRequest(
		WithMethod(http.MethodHead),
		WithQueryPath("/test"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/test", req.URL.String())
	assert.Equal(t, http.MethodHead, req.Method)
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
}
func TestCreteRequestWithUserMultiplePaths(t *testing.T) {
	req, err := createTestClient().NewRequest(
		WithMethod(http.MethodHead),
		WithQueryPath("/test"),
		WithQueryPath("tset"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/tset", req.URL.String())
	assert.Equal(t, http.MethodHead, req.Method)
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
}
func TestCreteRequestWithParams(t *testing.T) {
	req, err := createTestClient().NewRequest(
		WithMethod(http.MethodHead),
		WithQueryPath("/test"),
		WithQueryParam("k1", "v1"),
		WithQueryParam("k2", "v2"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/test?k1=v1&k2=v2", req.URL.String())
	assert.Equal(t, http.MethodHead, req.Method)
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
}
func TestCreteRequestWithMultiParams(t *testing.T) {
	req, err := createTestClient().NewRequest(
		WithMethod(http.MethodHead),
		WithQueryPath("/test"),
		WithQueryParam("k1", "v1"),
		WithQueryParam("k1", "v2"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/test?k1=v1&k1=v2", req.URL.String())
	assert.Equal(t, http.MethodHead, req.Method)
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
}
func TestCreteRequestWithArrayParams(t *testing.T) {
	req, err := createTestClient().NewRequest(
		WithMethod(http.MethodHead),
		WithQueryPath("/test"),
		WithQueryParamList("k1", "v1", "v2", "v3"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://test.com/test?k1=v1%2Cv2%2Cv3", req.URL.String())
	assert.Equal(t, http.MethodHead, req.Method)
	assert.Equal(t, "test-agent", req.Header.Get("User-Agent"))
}
