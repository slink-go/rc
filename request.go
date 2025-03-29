package rc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type requestBuilder struct {
	method      string
	baseUrl     *url.URL
	queryPath   *string
	queryParams url.Values
	body        any
	userAgent   *string
}

func (rb *requestBuilder) build() (*http.Request, error) {

	var u *url.URL
	var err error

	if rb.queryPath != nil {
		u, err = rb.baseUrl.Parse(strings.TrimPrefix(*rb.queryPath, "/"))
		if err != nil {
			return nil, fmt.Errorf("could not parse URL: %w", err)
		}
	} else {
		u = rb.baseUrl
	}
	//fmt.Println("<<<<<<<", u)

	var buf io.ReadWriter
	if rb.body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(rb.body)
		if err != nil {
			return nil, fmt.Errorf("could not serialize body: %w", err)
		}
	}

	urlStr := u.String()
	if rb.queryParams != nil && len(rb.queryParams) > 0 {
		//fmt.Println(">>>>>>>", rb.queryParams.Encode())
		urlStr += "?" + rb.queryParams.Encode()
	}

	req, err := http.NewRequest(rb.method, urlStr, buf)
	if err != nil {
		return nil, err
	}

	if rb.body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if rb.userAgent != nil && len(*rb.userAgent) > 0 {
		req.Header.Set("User-Agent", *rb.userAgent)
	}

	return req, nil

}
