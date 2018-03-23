package main

import (
	"fmt"
	"sync"
)

//
// Several solutions to the crawler exercise from the Go tutorial
// https://tour.golang.org/concurrency/10
//

//
// Serial crawler
//

// Serial 是按照顺序抓取
func Serial(url string, fetcher Fetcher, fetched map[string]bool) {
	// 如果抓取过，就直接停止程序
	if fetched[url] {
		return
	}
	// 标记 url 的状态为已抓取
	fetched[url] = true

	// 利用 fetcher.Fetch 抓取 url 页面中的链接
	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}

	// 再依次抓取 urls 中的网页
	for _, u := range urls {
		Serial(u, fetcher, fetched)
	}

	return
}

//
// Concurrent crawler with shared state and Mutex
//

// fetchState 管理抓取时候的状态
type fetchState struct {
	mu      sync.Mutex
	fetched map[string]bool
}

// ConcurrentMutex 并行互斥地抓取
func ConcurrentMutex(url string, fetcher Fetcher, f *fetchState) {
	// 在检查 url 的抓取状态时，先上互斥锁。
	// 防止数据竞争
	f.mu.Lock()
	// url 已经被抓取过了以后，直接结束程序
	if f.fetched[url] {
		f.mu.Unlock()
		return
	}

	// 没有抓取过 url 的话，就标记为已抓取
	f.fetched[url] = true
	f.mu.Unlock()

	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}

	// done 管理了爬取 urls 的任务群组
	var done sync.WaitGroup
	for _, u := range urls {
		// 给 done 多一份等待
		done.Add(1)
		go func(u string) {
			// 任务完成后，少一份等待
			defer done.Done()
			// 爬取 u
			ConcurrentMutex(u, fetcher, f)
		}(u)
	}

	// 等待 urls 中的网址全部爬取完成
	done.Wait()
	return
}

// 创建 *fetchState 实体
func makeState() *fetchState {
	f := &fetchState{}
	f.fetched = make(map[string]bool)
	return f
}

//
// Concurrent crawler with channels
// 使用通道来完成并行爬虫
//

// worker 爬取 url 页面后，把结果发送到 ch
// 如果爬取失败，就发送一个空切片到 ch
func worker(url string, ch chan []string, fetcher Fetcher) {
	urls, err := fetcher.Fetch(url)
	if err != nil {
		ch <- []string{}
	} else {
		ch <- urls
	}
}

// master 负责分配爬取任务给 worker
func master(ch chan []string, fetcher Fetcher) {
	n := 1
	fetched := make(map[string]bool)
	for urls := range ch {
		for _, u := range urls {
			if fetched[u] == false {
				fetched[u] = true
				n++
				go worker(u, ch, fetcher)
			}
		}
		n--
		if n == 0 {
			break
		}
	}
}

// ConcurrentChannel 启动 master
func ConcurrentChannel(url string, fetcher Fetcher) {
	ch := make(chan []string)
	go func() {
		ch <- []string{url}
	}()
	master(ch, fetcher)
}

//
// main
//

func main() {
	fmt.Printf("=== Serial===\n")
	Serial("http://golang.org/", fetcher, make(map[string]bool))

	fmt.Printf("=== ConcurrentMutex ===\n")
	ConcurrentMutex("http://golang.org/", fetcher, makeState())

	fmt.Printf("=== ConcurrentChannel ===\n")
	ConcurrentChannel("http://golang.org/", fetcher)
}

//
// Fetcher
//

// Fetcher 接口
type Fetcher interface {
	// Fetch returns a slice of URLs found on the page.
	Fetch(url string) (urls []string, err error)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) ([]string, error) {
	if res, ok := f[url]; ok {
		fmt.Printf("found:   %s\n", url)
		return res.urls, nil
	}
	fmt.Printf("missing: %s\n", url)
	return nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
