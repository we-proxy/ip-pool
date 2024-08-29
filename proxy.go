package ippool

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// proxy format: `scheme://hostname:port`
func Proxy(req *http.Request, proxy string, timeout time.Duration) (*http.Response, error) {
	segs := strings.SplitN(proxy, delimProtocol, 2)
	// boundary check, avoid out of range
	if len(segs) < 2 {
		return nil, fmt.Errorf("failed to parse proxy=%q", proxy)
	}
	scheme, host := segs[0], segs[1]
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{Scheme: scheme, Host: host}),
		},
	}
	if timeout > 0 {
		client.Timeout = timeout
	}
	resp, err := client.Do(req)
	if err != nil {
		// resp body should have been closed internally
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	contentLength, _ := strconv.Atoi(resp.Header.Get("content-length"))
	isProbablyProxyFailed := resp.StatusCode >= 500 ||
		resp.StatusCode == 403 ||
		resp.StatusCode == 406 ||
		contentLength <= 3
	if isProbablyProxyFailed {
		resp.Body.Close() // close for user if error
		return resp, fmt.Errorf("statusCode=%d, contentLength=%d", resp.StatusCode, contentLength)
	}
	return resp, nil
}
