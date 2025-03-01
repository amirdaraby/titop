package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/amirdaraby/titop/internal/usage"
)

func main() {
	cpuRes := make(chan usage.CPU, 1)
	memRes := make(chan usage.Memory, 1)

	for {
		go usage.Overall(cpuRes, memRes)

		clear := exec.Command("clear")
		clear.Stdout = os.Stdout

		err := clear.Run()

		if err != nil {
			panic(err)
		}

		cpuOverall := <-cpuRes
		memOverall := <-memRes

		fmt.Printf("UPTIME: %s\n", cpuOverall.UpTime.String())
		fmt.Printf("CPU USAGE: %.2f%%\n\n", cpuOverall.Usage)

		fmt.Println("CORES")
		for coreNum, core := range cpuOverall.Cores {
			fmt.Printf("CORE(%d) USAGE: %.2f%%\n", coreNum, core.Usage)
		}

		fmt.Printf("\n")

		fmt.Printf("MEMORY USAGE: %d/%d MB (%.2f%%)\n ", memOverall.Allocated/1000, memOverall.Total/1000, memOverall.Usage)

		time.Sleep(time.Second)
	}
}
