package main

import (
	"io"
	"log"
	"net/http"
	"time"

	ippool "github.com/we-proxy/ip-pool"
)

// const random = false
const random = true

// const concurrent = 10
const concurrent = 20

// const eachTimeout = 200 * time.Millisecond
const eachTimeout = 10 * time.Second

func main() {
	// See: https://github.com/Zaeem20/FREE_PROXIES_LIST
	// proxies, err := Load("https", "../FREE_PROXIES_LIST/https.txt") // 貌似全部阵亡
	proxies, err := ippool.Load("http", "../FREE_PROXIES_LIST/http.txt")
	if err != nil {
		log.Println("Failed to load pool:", err)
		return
	}
	if random {
		ippool.Shuffle(proxies)
	}
	req, err := http.NewRequest("GET", "http://ipinfo.io", nil)
	if err != nil {
		log.Println("Failed to create request:", err)
		return
	}
	resp, proxy, err := ippool.Race(req, proxies, concurrent, eachTimeout)
	if err != nil {
		log.Println("Failed to proxy request:", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read response:", err)
		return
	}
	log.Printf("Response from proxy %q: %s\n", proxy, string(body))
}
