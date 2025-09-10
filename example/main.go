package main

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	ippool "github.com/we-proxy/ip-pool"
)

// const random = false
const random = true

// const concurrent = 10
const concurrent = 20

// const eachTimeout = 200 * time.Millisecond
const eachTimeout = 10 * time.Second

const nOk = 5

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
	okItems, err := ippool.Race(req, proxies, concurrent, eachTimeout, nOk)
	if err != nil {
		log.Println("Failed to proxy request:", err)
		return
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
