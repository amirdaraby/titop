package application

import (
	"context"
	"time"

	"github.com/amirdaraby/titop/internal/usage"
)

func Run(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)

	cpuRes := make(chan usage.CPU, 1)
	memRes := make(chan usage.Memory, 1)

	ui, err := Init(cancel)

	if err != nil {
		return err
	}

	go ui.pollAndListenToEvents(cancel)

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			default:
				usage.Overall(cpuRes, memRes)
				time.Sleep(time.Second * 1)
			}
		}
	}()

	for {
		cpu := <-cpuRes
		mem := <-memRes

		ui.display(cpu, mem)
	}
}
