package ippool

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

// const random = false
const random = true

// const concurrent = 10
const concurrent = 20

// const eachTimeout = 200 * time.Millisecond
const eachTimeout = 10 * time.Second

const nOk = 5

func TestPool(t *testing.T) {
	// proxies, err := Load("https", "./FREE_PROXIES_LIST/https.txt") // 貌似全部阵亡
	proxies, err := Load("http", "./FREE_PROXIES_LIST/http.txt")
	if err != nil {
		t.Fatal("Failed to load pool:", err)
	}
	if random {
		Shuffle(proxies)
	}
	req, err := http.NewRequest("GET", "http://ipinfo.io", nil)
	if err != nil {
		t.Fatal("Failed to create request:", err)
	}
	okItems, err := Race(req, proxies, concurrent, eachTimeout, nOk)
	if err != nil {
		t.Fatal("Failed to proxy request:", err)
	}
	for i, item := range okItems {
		func() {
			defer item.Resp.Body.Close()
			log.Printf("------- okItem #%d -------\n", i+1)
			body, err := io.ReadAll(item.Resp.Body)
			if err != nil {
				log.Println("Failed to read response:", err)
				return
			}
			re := regexp.MustCompile(`(?i)error|html|doctype|login|password`)
			if re.Match(body) {
				log.Printf("Response from proxy %q: not a valid IP Info\n", item.Proxy)
				return
			}
			log.Printf("Response from proxy %q: %s\n", item.Proxy, body)
		}()
	}
}

func TestBoundaryCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://foo", nil)
	if _, err := Race(req, []string{"1.2.3.4"}, concurrent, eachTimeout, nOk); err == nil ||
		!strings.Contains(err.Error(), "failed to parse proxy") {
		t.Fatalf("err=%v should be %q", err, "failed to parse proxy")
	}
	if _, err := Race(req, []string{"", "", ""}, concurrent, eachTimeout, nOk); err == nil ||
		!strings.Contains(err.Error(), "failed to parse proxy") {
		t.Fatalf("err=%v should be %q", err, "failed to parse proxy")
	}
	if _, err := Race(req, []string{}, concurrent, eachTimeout, nOk); err == nil ||
		!strings.Contains(err.Error(), "be gte 1") {
		t.Fatalf("err=%v should be %q", err, "be gte 1")
	}
	if _, err := Race(req, []string{"1.2.3.4"}, 0, eachTimeout, nOk); err == nil ||
		!strings.Contains(err.Error(), "be gte 1") {
		t.Fatalf("err=%v should be %q", err, "be gte 1")
	}
	if _, err := Race(req, []string{"1.2.3.4"}, -1, eachTimeout, nOk); err == nil ||
		!strings.Contains(err.Error(), "be gte 1") {
		t.Fatalf("err=%v should be %q", err, "be gte 1")
	}
}
