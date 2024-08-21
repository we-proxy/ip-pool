package ippool

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type okItem struct {
	proxy string
	res   []byte
}
type errItem struct {
	proxy string
	err   error
}

const requestInterval = time.Second * 1

func Request(req *http.Request, proxies []string, concurrent int) ([]byte, string, error) {
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
				if res, err := requestProxy(req, proxy); err != nil {
					errCh <- errItem{proxy, err}
				} else {
					okCh <- okItem{proxy, res}
				}
			}(proxy)
		}
		var errCnt int
		select {
		case item := <-okCh:
			return item.res, item.proxy, nil
		case item := <-errCh:
			lastErrItem = item
			errCnt++
			if errCnt >= end-i {
				continue
			}
		}
		time.Sleep(requestInterval) // 防止短时间内发送太多请求
	}
	return nil, lastErrItem.proxy, lastErrItem.err
}

// proxy format: `scheme://hostname:port`
func requestProxy(req *http.Request, proxy string) ([]byte, error) {
	segs := strings.SplitN(proxy, delimProtocol, 2)
	scheme, host := segs[0], segs[1]
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{Scheme: scheme, Host: host}),
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("Status Code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %w", err)
	}
	return body, nil
}
