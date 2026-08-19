package hedge

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

func NewHedging() *HedgeT {
	h := &HedgeT{
		lock:       new(sync.RWMutex),
		delay:      10 * time.Second, // delay between requests
		maxRequest: 3,                // max requests
	}
	return h
}

type HedgeT struct {
	lock       *sync.RWMutex
	underlying http.RoundTripper
	delay      time.Duration
	maxRequest int
}

// SetTransport sets the underlying HTTP transport that [HedgeT] delegates
// individual requests to.
func (ht *HedgeT) SetTransport(t http.RoundTripper) *HedgeT {
	ht.lock.Lock()
	defer ht.lock.Unlock()
	ht.underlying = t
	return ht
}

// Transport returns the underlying HTTP transport.
func (ht *HedgeT) Transport() http.RoundTripper {
	ht.lock.RLock()
	defer ht.lock.RUnlock()
	return ht.underlying
}

// Delay method returns the delay between hedged requests.
func (ht *HedgeT) Delay() time.Duration {
	ht.lock.RLock()
	defer ht.lock.RUnlock()
	return ht.delay
}

// SetDelay method sets the delay between hedged requests.
func (ht *HedgeT) SetDelay(delay time.Duration) *HedgeT {
	ht.lock.Lock()
	defer ht.lock.Unlock()
	ht.delay = delay
	return ht
}

// MaxRequest method returns the maximum number of concurrent hedged requests.
func (ht *HedgeT) MaxRequest() int {
	ht.lock.RLock()
	defer ht.lock.RUnlock()
	return ht.maxRequest
}

// SetMaxRequest method sets the maximum number of concurrent hedged requests.
func (ht *HedgeT) SetMaxRequest(count int) *HedgeT {
	ht.lock.Lock()
	defer ht.lock.Unlock()
	if count <= 0 {
		ht.maxRequest = 1
	}
	ht.maxRequest = count
	return ht
}

func (ht *HedgeT) RoundTrip(req *http.Request) (*http.Response, error) {
	ht.lock.RLock()
	underlying := ht.underlying
	maxReq := ht.maxRequest
	delay := ht.delay
	ht.lock.RUnlock()

	if maxReq <= 1 {
		return underlying.RoundTrip(req)
	}

	reqCtx := req.Context()

	type result struct {
		resp *http.Response
		err  error
	}
	resultCh := make(chan result, 1)
	decidedCh := make(chan struct{})

	var (
		mu       sync.Mutex
		decided  bool
		inflight = make(map[int]context.CancelFunc, maxReq)
	)

	// decide records attempt i as the winner. It reports false when another
	// attempt already won. The winner's own cancel func is left untouched; every
	// other in-flight attempt is cancelled right away.
	decide := func(i int) bool {
		mu.Lock()
		if decided {
			mu.Unlock()
			return false
		}
		decided = true
		losers := make([]context.CancelFunc, 0, len(inflight))
		for idx, cancel := range inflight {
			if idx != i {
				losers = append(losers, cancel)
			}
		}
		clear(inflight)
		mu.Unlock()

		for _, cancel := range losers {
			cancel()
		}
		close(decidedCh)
		return true
	}

spawn:
	for i := range maxReq {
		if i > 0 {
			if delay > 0 {
				select {
				case <-time.After(delay):
				case <-decidedCh:
					break spawn
				case <-reqCtx.Done():
					break spawn
				}
			}
		}

		attemptCtx, attemptCancel := context.WithCancel(reqCtx)

		mu.Lock()
		if decided {
			mu.Unlock()
			attemptCancel()
			break spawn
		}
		inflight[i] = attemptCancel
		mu.Unlock()

		go func(i int, attemptCancel context.CancelFunc) {
			resp, err := underlying.RoundTrip(req.Clone(attemptCtx))

			if !decide(i) {
				attemptCancel()
				if resp != nil && resp.Body != nil {
					drainReadCloser(resp.Body)
				}
				return
			}

			// Winner: the caller still has to read this body, so hand the cancel
			// func over to it instead of firing it here.
			if resp != nil && resp.Body != nil {
				resp.Body = &cancelReadCloser{r: resp.Body, cancel: attemptCancel}
			} else {
				attemptCancel()
			}
			resultCh <- result{resp: resp, err: err}
		}(i, attemptCancel)
	}

	res := <-resultCh
	return res.resp, res.err
}

type cancelReadCloser struct {
	r      io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelReadCloser) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c *cancelReadCloser) Close() error {
	err := c.r.Close()
	c.cancel()
	return err
}

func closeq(v any) {
	if c, ok := v.(io.Closer); ok {
		silently(c.Close())
	}
}
func silently(_ ...any) {}

func drainReadCloser(body io.ReadCloser) {
	if body != nil {
		defer closeq(body)
		_, _ = io.Copy(io.Discard, body)
	}
}
