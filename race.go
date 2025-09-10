package ippool

import (
	"fmt"
	"net/http"
	"time"
)

type okItem struct {
	Resp  *http.Response
	Proxy string
}

type errItem struct {
	resp  *http.Response // Note: resp might be nil if error occurred before getting a response
	proxy string
	err   error
}

const requestInterval = time.Second * 1

// RaceModified modifies the original Race function to collect at least minOkItems successful responses
// before returning. It returns a slice of okItem if successful, or the last error if not enough items
// could be collected.
func Race(req *http.Request, proxies []string, concurrent int, eachTimeout time.Duration, minOkItems int) ([]okItem, error) {
	// Boundary checks
	if len(proxies) < 1 {
		return nil, fmt.Errorf("len(proxies)=%d should be gte 1", len(proxies))
	}
	if concurrent < 1 {
		return nil, fmt.Errorf("concurrent=%d should be gte 1", concurrent)
	}
	if minOkItems < 1 {
		return nil, fmt.Errorf("minOkItems=%d should be gte 1", minOkItems)
	}

	// Channels for communication
	okCh := make(chan okItem, len(proxies))   // Buffered to prevent goroutine leaks if not read
	errCh := make(chan errItem, len(proxies)) // Buffered to prevent goroutine leaks if not read

	totalProxies := len(proxies)
	collectedOkItems := make([]okItem, 0, minOkItems) // Pre-allocate capacity for efficiency
	var lastErrItem errItem
	sentRequests := 0
	receivedResponses := 0

	// sendRequests sends requests for a batch of proxies
	sendRequests := func(startIdx, endIdx int) {
		for i := startIdx; i < endIdx; i++ {
			proxy := proxies[i]
			sentRequests++
			go func(p string) {
				resp, err := Proxy(req, p, eachTimeout)
				if err != nil {
					errCh <- errItem{resp: resp, proxy: p, err: err}
				} else {
					okCh <- okItem{Resp: resp, Proxy: p}
				}
			}(proxy)
		}
	}

	// Send initial batch
	initialBatchEnd := min(concurrent, totalProxies)
	sendRequests(0, initialBatchEnd)
	currentIndex := initialBatchEnd // Index for the next proxy to send

	// Main loop: collect results until we have enough okItems or determine we can't get enough
	for receivedResponses < totalProxies && len(collectedOkItems) < minOkItems {
		// Calculate how many more requests we can send in the next batch
		// This is a simple approach, you might want a more sophisticated rate limiting
		potentialNextBatchEnd := min(currentIndex+concurrent, totalProxies)
		requestsNeeded := potentialNextBatchEnd - currentIndex

		// Decide if we should wait for results or send more requests
		// A simple heuristic: send more if we haven't sent all and have capacity in our goal vs sent ratio
		shouldSendMore := currentIndex < totalProxies && (len(collectedOkItems)+requestsNeeded) <= minOkItems+concurrent // Avoid sending way too many

		if shouldSendMore && requestsNeeded > 0 {
			sendRequests(currentIndex, potentialNextBatchEnd)
			currentIndex = potentialNextBatchEnd
		}

		// Wait for a result (ok or error)
		select {
		case item := <-okCh:
			receivedResponses++
			collectedOkItems = append(collectedOkItems, item)
			// Check if we have collected enough
			if len(collectedOkItems) >= minOkItems {
				// Success: we have enough items
				return collectedOkItems, nil
			}
		case item := <-errCh:
			receivedResponses++
			lastErrItem = item // Keep track of the last error
			// Optional: Log error or handle specific errors if needed
		}

		// If we've sent all requests and received all responses but still don't have enough
		if sentRequests == totalProxies && receivedResponses == totalProxies && len(collectedOkItems) < minOkItems {
			// Failure: Not enough successful items
			// Return the items we have (could be empty) and the last error encountered
			// You might want to return a specific error indicating "not enough items"
			if len(collectedOkItems) > 0 {
				return collectedOkItems, fmt.Errorf("only collected %d ok items, needed %d. Last error: %w", len(collectedOkItems), minOkItems, lastErrItem.err)
			} else {
				// If no successful items, return the last error
				return nil, fmt.Errorf("failed to collect any ok items, needed %d. Last error: %w", minOkItems, lastErrItem.err)
			}
		}

		// Small delay to prevent busy waiting and respect requestInterval
		// Note: This simplistic delay might not be perfect for rate limiting across batches.
		// Consider a more robust rate limiter if precise timing is needed.
		time.Sleep(requestInterval / 10) // Smaller sleep inside the loop
	}

	// This point might be reached if loops end without explicit return (e.g., all sent, waiting)
	// Should ideally be covered by the condition inside the loop.
	// If we exit loop without returning, it implies we either got enough or ran out of possibilities.
	if len(collectedOkItems) >= minOkItems {
		return collectedOkItems, nil
	}
	// If not enough, it should have been caught in the loop. Defensive programming.
	return nil, fmt.Errorf("race modified loop ended without sufficient items or clear error, collected %d, needed %d", len(collectedOkItems), minOkItems)
}
