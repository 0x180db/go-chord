package chord

import (
	"context"

	"github.com/0x180db/go-conduit"
)

type Flow[In, Out any] interface {
	OnSuccess(context.Context, Out) error
	OnError(context.Context, error)
	Pipeline(Stage[In]) Stage[Out]
}

type Stage[T any] conduit.Stage[T]

func NewStage[In, Out any](p Stage[In], fn func(context.Context, In) (Out, error)) Stage[Out] {
	return Stage[Out](
		conduit.NewProducerConsumer(conduit.Stage[In](p), fn),
	)
}

type Trigger[T any] interface {
	Stage(context.Context) Stage[T]
}

func RunFlow[In, Out any](s Stage[In], f Flow[In, Out]) {
	conduit.NewConsumer(
		conduit.Stage[Out](f.Pipeline(s)),
		f.OnSuccess,
		f.OnError,
	)
}
