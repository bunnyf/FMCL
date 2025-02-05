package datafetcher

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Fetcher struct {
	client   *http.Client
	interval time.Duration
}

func NewFetcher(interval time.Duration) *Fetcher {
	return &Fetcher{
		client:   &http.Client{Timeout: 10 * time.Second},
		interval: interval,
	}
}

func (f *Fetcher) Run(ctx context.Context, url string) <-chan string {
	dataChan := make(chan string, 10)
	go func() {
		ticker := time.NewTicker(f.interval)
		defer close(dataChan)
		for {
			select {
			case <-ticker.C:
				go func() {
					resp, err := f.client.Get(url)
					if err != nil {
						return
					}
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)
					dataChan <- string(body)
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
	return dataChan
}
