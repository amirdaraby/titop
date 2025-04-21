package collect

import (
	"github.com/amirdaraby/titop/internal/collect/cpu"
	"github.com/amirdaraby/titop/internal/collect/mem"
	"github.com/amirdaraby/titop/internal/collect/proc"
)

func Collect(cpuRes chan cpu.CPU, memRes chan mem.Memory, processesRes chan []proc.Process) {
	go cpu.SendUsage(cpuRes)
	go mem.SendUsage(memRes)
	go proc.SendUsage(processesRes)
}