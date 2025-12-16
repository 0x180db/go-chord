# go-chord
Event-driven workflow engine built on go-conduit

## Overview

go-chord provides a high-level abstraction for building event-driven workflows in Go. It sits on top of [go-conduit](https://github.com/0x180db/go-conduit), adding workflow composition and trigger-based execution patterns.

## Features

- **Event-Driven**: Trigger-based workflow execution (timers, webhooks, queues)
- **Stream Processing**: Stages process data as soon as they're free - no waiting for batches
- **Railway Oriented**: Automatic error propagation through success/error paths
- **Type Safe**: Leverage Go generics for compile-time safety
- **Dependency Injection**: Clean architecture with testable workflows

## Installation
```bash
go get github.com/0x180db/go-chord
```

## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/0x180db/go-chord"
    "github.com/0x180db/go-chord/trigger"
)

// Define your workflow
type PrinterFlow struct{}

func (f PrinterFlow) OnSuccess(ctx context.Context, msg string) error {
    fmt.Println("Message:", msg)
    return nil
}

func (f PrinterFlow) OnError(ctx context.Context, err error) {
    fmt.Println("Error:", err)
}

func (f PrinterFlow) Pipeline(s chord.Stage[time.Time]) chord.Stage[string] {
    return chord.NewStage(s, func(ctx context.Context, t time.Time) (string, error) {
        return fmt.Sprintf("Tick at %s", t.Format(time.RFC3339)), nil
    })
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Create a ticker trigger that fires every second
    ticker := trigger.NewTicker(time.Second)
    
    // Run the workflow
    chord.RunFlow(ticker.Stage(ctx), PrinterFlow{})
}
```

## Core Components

### Flow Interface

The `Flow` interface defines how your workflow processes data:
```go
type Flow[In, Out any] interface {
    // Success handler - called for each successful result
    OnSuccess(context.Context, Out) error
    
    // Error handler - called when any stage returns an error
    OnError(context.Context, error)
    
    // Pipeline - defines the data transformation stages
    Pipeline(Stage[In]) Stage[Out]
}
```

### Building Pipelines

Chain multiple stages to transform data:
```go
func (f MyFlow) Pipeline(s chord.Stage[Input]) chord.Stage[Output] {
    // Stage 1: Input -> Intermediate
    stage1 := chord.NewStage(s, func(ctx context.Context, in Input) (Intermediate, error) {
        // transform input
        return intermediate, nil
    })
    
    // Stage 2: Intermediate -> Output
    stage2 := chord.NewStage(stage1, func(ctx context.Context, mid Intermediate) (Output, error) {
        // transform intermediate
        return output, nil
    })
    
    return stage2
}
```

### Triggers

Triggers generate events that start your workflows:
```go
type Trigger[T any] interface {
    Stage(context.Context) chord.Stage[T]
}
```

**Built-in triggers:**
- `trigger.NewTicker(duration)` - fires at regular intervals

**Custom trigger example:**
```go
type WebhookTrigger struct {
    addr string
}

func (t WebhookTrigger) Stage(ctx context.Context) chord.Stage[WebhookEvent] {
    return func() <-chan conduit.Result[WebhookEvent] {
        ch := make(chan conduit.Result[WebhookEvent])
        
        go func() {
            defer close(ch)
            
            http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
                event := parseWebhook(r)
                ch <- conduit.Ok(ctx, event)
            })
            
            http.ListenAndServe(t.addr, nil)
        }()
        
        return ch
    }
}
```

## Error Handling

Errors flow through the pipeline automatically to your `OnError` handler:
```go
func (f MyFlow) Pipeline(s chord.Stage[Input]) chord.Stage[Output] {
    return chord.NewStage(s, func(ctx context.Context, in Input) (Output, error) {
        if in.IsInvalid() {
            // Error flows to OnError handler
            return Output{}, errors.New("invalid input")
        }
        return process(in), nil
    })
}

func (f MyFlow) OnError(ctx context.Context, err error) {
    // Handle all pipeline errors here
    log.Printf("Pipeline error: %v", err)
}
```

## Context Cancellation

Workflows respect context cancellation for graceful shutdown:
```go
ctx, cancel := context.WithCancel(context.Background())

// Start workflow
go chord.RunFlow(trigger.Stage(ctx), flow)

// Cancel when done
cancel()
```

## Best Practices

- **Inject dependencies into Flow structs** - makes testing easier and keeps flows pure
- **Keep stages focused** - each stage should do one transformation
- **Handle errors in OnError** - centralized error handling prevents scattered error logic
- **Use context for cancellation** - enables graceful shutdown
- **Make Flows stateless** - store state in injected dependencies, not in the Flow itself

## License

MIT

## Related Projects

- [go-conduit](https://github.com/0x180db/go-conduit) - Concurrent and multi-stage pipelines in Go
