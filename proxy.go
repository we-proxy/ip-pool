package ippool

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type okItem struct {
	resp  *http.Response
	proxy string
}
type errItem struct {
	resp  *http.Response
	proxy string
	err   error
}

const requestInterval = time.Second * 1

func Request(req *http.Request, proxies []string, concurrent int) (*http.Response, string, error) {
	// boundary check, avoid infinite loop
	if len(proxies) < 1 {
		return nil, "", fmt.Errorf("len(proxies)=%d should be gte 1", len(proxies))
	}
	if concurrent < 1 {
		return nil, "", fmt.Errorf("concurrent=%d should be gte 1", concurrent)
	}
	okCh := make(chan okItem)
	errCh := make(chan errItem)
	var lastErrItem errItem

	for i := 0; i < len(proxies); i += concurrent {
		end := i + concurrent
		if end > len(proxies) {
			end = len(proxies)
		}
		for _, proxy := range proxies[i:end] {
			go func(proxy string) {
				if resp, err := requestProxy(req, proxy); err != nil {
					errCh <- errItem{resp, proxy, err}
				} else {
					okCh <- okItem{resp, proxy}
				}
			}(proxy)
		}
		var errCnt int
		select {
		case item := <-okCh:
			return item.resp, item.proxy, nil
		case item := <-errCh:
			lastErrItem = item
			errCnt++
			if errCnt >= end-i {
				continue
			}
		}
		time.Sleep(requestInterval) // 防止短时间内发送太多请求
	}
	return lastErrItem.resp, lastErrItem.proxy, lastErrItem.err
}

// proxy format: `scheme://hostname:port`
func requestProxy(req *http.Request, proxy string) (*http.Response, error) {
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
