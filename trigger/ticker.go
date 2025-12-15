package trigger

import (
	"context"
	"time"

	"github.com/0x180db/go-chord"
	"github.com/0x180db/go-conduit"
)

type Ticker struct {
	*time.Ticker
}

func NewTicker(d time.Duration) chord.Trigger[time.Time] {
	return Ticker{time.NewTicker(time.Second)}
}

func (t Ticker) Stage(ctx context.Context) chord.Stage[time.Time] {
	return func() <-chan conduit.Result[time.Time] {
		ch := make(chan conduit.Result[time.Time])

		go func() {
			defer t.Stop()
			defer close(ch)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					ch <- conduit.Ok(ctx, <-t.C)
				}
			}
		}()

		return ch
	}
}
