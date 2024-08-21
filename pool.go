package ippool

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// scheme=`http`, filename=`./FREE_PROXIES_LIST/http.txt`
func LoadPool(scheme, filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var proxies []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			proxy := fmt.Sprintf("%s%s%s", scheme, delimProtocol, line)
			proxies = append(proxies, proxy)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error reading file: %w", err)
	}
	return proxies, nil
}
