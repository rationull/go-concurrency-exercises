//////////////////////////////////////////////////////////////////////
//
// Your task is to change the code to limit the crawler to at most one
// page per second, while maintaining concurrency (in other words,
// Crawl() must be called concurrently)
//
// @hint: you can achieve this by adding 3 lines
// JCL: I opted for a > 3 line solution to synchronize fetch calls with a channel
// rather than just doing a synchronized sleep in Crawl(). Synchronizing around
// the Fetch() call with a mutex and keeping a lastFetchTime there is an alternative
// requiring fewer code changes, but IMO the channel solution is a more interesting
// exercise from the perspective of practicing with Go's concurrency primitives.
//

package main

import (
	"fmt"
	"sync"
	"time"
)

type fetchResult struct {
	body string
	urls []string
	err  error
}

// Crawl uses `fetcher` from the `mockfetcher.go` file to imitate a
// real crawler. It crawls until the maximum depth has reached.
func Crawl(url string, depth int, wg *sync.WaitGroup, fetchChannel chan string, resultChannel chan fetchResult) {
	defer wg.Done()

	if depth <= 0 {
		return
	}

	fetchChannel <- url
	res := <-resultChannel
	body, urls, err := res.body, res.urls, res.err
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("found: %s %q\n", url, body)

	wg.Add(len(urls))
	for _, u := range urls {
		// Do not remove the `go` keyword, as Crawl() must be
		// called concurrently
		go Crawl(u, depth-1, wg, fetchChannel, resultChannel)
	}
	return
}

func doFetches(fetchChannel chan string, resultChannel chan fetchResult) {
	lastFetch := time.Now()

	for {
		url := <-fetchChannel

		elapsed := time.Since(lastFetch)
		if elapsed < 1*time.Second {
			time.Sleep(1*time.Second - elapsed)
		}

		body, urls, err := fetcher.Fetch(url)
		lastFetch = time.Now()

		resultChannel <- fetchResult{body, urls, err}
	}
}

func main() {
	var wg sync.WaitGroup
	fetchChannel := make(chan string)
	resultChannel := make(chan fetchResult)

	go doFetches(fetchChannel, resultChannel)

	wg.Add(1)
	Crawl("http://golang.org/", 4, &wg, fetchChannel, resultChannel)
	wg.Wait()
}
