package serverselect

import (
	"context"
	"log"
	"net/http"
	"sync"
)

func Racer(ctx context.Context, a string, b string) string {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	aNotify := ping(ctx, a)
	bNotify := ping(ctx, b)

	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	case <-aNotify:
		return a
	case <-bNotify:
		return b
	}
}

func ping(ctx context.Context, url string) chan struct{} {
	ch := make(chan struct{})

	go func() {
		defer close(ch)

		log.Printf("getting: %q\n", url)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			log.Printf("request error for %q: %q", url, err.Error())
			return
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("error: %q", err.Error())
			return
		}

		if res.StatusCode == http.StatusOK {
			log.Printf("returning: %q\n", url)
			ch <- struct{}{}
		}
	}()

	return ch
}

func RacerVar(ctx context.Context, urls ...string) string {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	aggr := make(chan string, len(urls))
	defer close(aggr)

	for _, url := range urls {
		c := pingVar(ctx, url)

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					log.Printf("stopping: %q", ctx.Err().Error())
					return
				case url := <-c:
					aggr <- url
					return
				}
			}
		}(&wg)
	}

	select {
	case <-ctx.Done():
		return ctx.Err().Error()
	case v := <-aggr:
		wg.Wait()
		return v
	}
}

func pingVar(ctx context.Context, url string) chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}

		if res.StatusCode == http.StatusOK {
			ch <- url
		}
	}()

	return ch
}
