# ip-pool

See also: https://github.com/Zaeem20/FREE_PROXIES_LIST

## Running Test

```sh
git clone git@github.com:we-proxy/ip-pool.git
cd ip-pool
# clone FREE_PROXIES_LIST to ./ (git-history too large, not recommended)
# git clone git@github.com/Zaeem20/FREE_PROXIES_LIST.git
# or download FREE_PROXIES_LIST (recommended)
mkdir FREE_PROXIES_LIST
curl https://fastly.jsdelivr.net/gh/Zaeem20/FREE_PROXIES_LIST@master/http.txt > FREE_PROXIES_LIST/http.txt
go test
>> ...
PASS
ok  	github.com/we-proxy/ip-pool	0.596s
```

## Running Example

```sh
cd example
go run .
>> Response from proxy "http://135.181.154.225:80": {
  "ip": "135.181.154.225",
  "hostname": "repo.getlic.pro",
  "city": "Helsinki",
  "region": "Uusimaa",
  "country": "FI",
  "loc": "60.1695,24.9354",
  "org": "AS24940 Hetzner Online GmbH",
  "postal": "00100",
  "timezone": "Europe/Helsinki",
  "readme": "https://ipinfo.io/missingauth"
}
```

## Import and Use

```go
import ippool "github.com/we-proxy/ip-pool"
// ...
const random = true
const concurrent = 10

func main() {
	// See: https://github.com/Zaeem20/FREE_PROXIES_LIST
	// proxies, err := ippool.LoadPool("https", "path/to/https.txt") // 貌似全部阵亡
	proxies, err := ippool.LoadPool("http", "path/to/http.txt")
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
	resp, proxy, err := ippool.Request(req, proxies, concurrent)
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
	log.Printf("Response from proxy %q: %s\n", proxy, string(res))
}
```
