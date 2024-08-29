package ippool

import (
	"fmt"
	"net/http"
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

func Race(req *http.Request, proxies []string, concurrent int, eachTimeout time.Duration) (*http.Response, string, error) {
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
				if resp, err := Proxy(req, proxy, eachTimeout); err != nil {
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
