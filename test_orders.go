//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	const numOrders = 100
	const baseURL = "http://localhost:8080/orders"

	var wg sync.WaitGroup
	results := make(chan string, numOrders)

	start := time.Now()

	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req, _ := http.NewRequest("POST", baseURL, nil)
			req.Header.Set("Idempotency-Key", fmt.Sprintf("test-order-%d", id))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- fmt.Sprintf("[%d] ERROR: %v", id, err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			results <- fmt.Sprintf("[%d] %d: %s", id, resp.StatusCode, string(body))
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	success := 0
	failed := 0
	for r := range results {
		fmt.Println(r)
		if len(r) > 0 && r[len(r)-1] == '}' {
			success++
		} else {
			failed++
		}
	}

	elapsed := time.Since(start)
	fmt.Println("\n--- Summary ---")
	fmt.Printf("Total: %d orders\n", numOrders)
	fmt.Printf("Success: %d\n", success)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Duration: %s\n", elapsed)
	fmt.Printf("Rate: %.2f orders/sec\n", float64(numOrders)/elapsed.Seconds())
}
