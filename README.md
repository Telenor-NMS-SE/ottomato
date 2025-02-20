# Ottomato - infrastructure automation scheduler

## What does it do?
Ottomoato is a framework to help you abstract task scheduling and automations accross your infrastructure with a BYO-X approach.
- Bring your own platform integration
- Bring your own state storage
- Bring your own transport

## What motivated us to make it?
Throughout the years we've used a few other frameworks and tools that let's you orchestrate a swarm of OT devices, but as we've been scaling up the project, we've encoutered some issues with availability zone, resource and orchestration management when dealing with 10-100k+ devices.

## Quick start
`go get github.com/telenor-nms-se/ottomato@0.0.1`

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	
	"github.com/Telenor-NMS-SE/ottomato/store"
	"github.com/Telenor-NMS-SE/ottomato/worker"
)

type ExampleDevice struct {
	Hostname string
}

func (d *ExampleDevice) Init(ctx context.Context) error {
	slog.Info("i have been initated", "device", d.Hostname)
	return nil
}

func (d *ExampleDevice) Ping(ctx context.Context) error {
	slog.Info("i have been pinged", "device", d.Hostname)
	return nil
}

func (d *ExampleDevice) RunTask(ctx context.Context, target string, task *worker.Task) (worker.Result, error) {
	slog.Info("i have received a task", "device", d.Hostname)
	return worker.Result{}, nil
}

func (d *ExampleDevice) Stop() error {
	slog.Info("i have been told to stop", "device", d.Hostname)
	return nil
}

func (d *ExampleDevice) Name() string {
	return d.Hostname
}

func (d *ExampleDevice) Info() map[string]any {
	return map[string]any{}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
 
	w, err := worker.New(ctx, store.New(ctx))
	if err != nil {
		slog.Error("failed to create new worker", "error", err)
	}
	defer func() {
		if err := w.Stop(); err != nil {
			slog.Error("failed to stop worker", "error", err)
		}
	}()

	for i := range 10 {
        name := fmt.Sprintf("device-%d", i)

        slog.Info("adding workload", "name", name)

		if err := w.AddWorkload(ctx, &ExampleDevice{Hostname: name}); err != nil {
			slog.Error("failed to add workload", "error", err)
		}
	}

	for _, wl := range w.Workloads() {
		res, err := w.RunTask(ctx, wl, &worker.Task{Command: "hello"})
		if err != nil {
			slog.Error("failed to run task", "workload", wl, "error", err)
        }
		slog.Info("got response", "workload", wl, "response", res)
        
        if err := w.DeleteWorkload(wl); err != nil {
            slog.Error("failed to delete workload", "workload", wl, "error", err)
        }
	}

	<-ctx.Done()
}
```

## Challenges
- Naming and interfaces
- Implementation of Manager
- Scope

## Roadmap
It is on the roadmap that Ottomato will be a multi-tierd orchestrator, where a Manager orchestrates Workers to handle scaling and multi-AZ better.

Improve scheduling when distributing tasks from a worker to a workload.

## License
Apache License 2.0