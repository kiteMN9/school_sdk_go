package client

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func (a *APIClient) save(cfg *APIConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	defer close(sigCh)
	go func() {
		defer close(done)
		a.getPubParams(ctx, cfg, true)
		if ctx.Err() != nil {
			return
		}
		a.getCourseListPre(ctx, cfg, cfg.xkkz_id, cfg.xszxzt, true)
	}()
	select {
	case <-sigCh:
		cancel()
		// fmt.Println("<-sigCh")
	case <-done:
		// fmt.Println("<-done")
		cancel()
	}
}
