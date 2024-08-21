package ippool

import (
	"io"
	"log"
	"net/http"
	"testing"
)

// const random = false
const random = true

// const concurrent = 10
const concurrent = 20

func TestPool(t *testing.T) {
	// proxies, err := LoadPool("https", "./FREE_PROXIES_LIST/https.txt") // 貌似全部阵亡
	proxies, err := LoadPool("http", "./FREE_PROXIES_LIST/http.txt")
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
	resp, proxy, err := Request(req, proxies, concurrent)
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
