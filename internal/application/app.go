package application

import (
	"context"
	"time"

	"github.com/amirdaraby/titop/internal/collect"
	"github.com/amirdaraby/titop/internal/collect/cpu"
	"github.com/amirdaraby/titop/internal/collect/mem"
	"github.com/amirdaraby/titop/internal/collect/proc"
	"github.com/amirdaraby/titop/internal/shared"
)

func Run(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)

	cpuRes := make(chan cpu.CPU, 1)
	memRes := make(chan mem.Memory, 1)
	processesRes := make(chan []proc.Process, 1)

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
				shared.Refreshing()
				collect.Collect(cpuRes, memRes, processesRes)
				time.Sleep(time.Millisecond * time.Duration(shared.GetRefreshRate()))
			}
		}
	}()

	for {
		cpu := <-cpuRes
		mem := <-memRes
		proc := <-processesRes
		

		ui.update(cpu, mem, proc)
	}
}
