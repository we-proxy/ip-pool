package ippool

import (
	"io"
	"log"
	"net/http"
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
	resp, proxy, err := Race(req, proxies, concurrent, eachTimeout)
	if err != nil {
		t.Fatal("Failed to proxy request:", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Failed to read response:", err)
	}
	log.Printf("Response from proxy %q: %s\n", proxy, string(body))
}

func TestBoundaryCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://foo", nil)
	if _, _, err := Race(req, []string{"1.2.3.4"}, concurrent, eachTimeout); err == nil ||
		!strings.Contains(err.Error(), "failed to parse proxy") {
		t.Fatalf("err=%v should be %q", err, "failed to parse proxy")
	}
	if _, _, err := Race(req, []string{"", "", ""}, concurrent, eachTimeout); err == nil ||
		!strings.Contains(err.Error(), "failed to parse proxy") {
		t.Fatalf("err=%v should be %q", err, "failed to parse proxy")
	}
	if _, _, err := Race(req, []string{}, concurrent, eachTimeout); err == nil ||
		!strings.Contains(err.Error(), "be gte 1") {
		t.Fatalf("err=%v should be %q", err, "be gte 1")
	}
	if _, _, err := Race(req, []string{"1.2.3.4"}, 0, eachTimeout); err == nil ||
		!strings.Contains(err.Error(), "be gte 1") {
		t.Fatalf("err=%v should be %q", err, "be gte 1")
	}
	if _, _, err := Race(req, []string{"1.2.3.4"}, -1, eachTimeout); err == nil ||
		!strings.Contains(err.Error(), "be gte 1") {
		t.Fatalf("err=%v should be %q", err, "be gte 1")
	}
}
