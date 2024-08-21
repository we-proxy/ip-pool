package ippool

import (
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
	res, proxy, err := Request(req, proxies, concurrent)
	if err != nil {
		t.Fatal("Failed to proxy request:", err)
	}
	log.Printf("Response from proxy %q: %s\n", proxy, string(res))
}
