package mem

import (
	"strconv"
	"strings"

	"github.com/amirdaraby/titop/internal/reader"
)

type Memory struct {
	Usage                       float32
	Total, Available, Allocated int // KB
	Swap                        *Memory
}

const (
	VM_SIZE_MEM = 16
	VM_RSS_MEM  = 20
	VM_SWAP_MEM = 25
)

func SendUsage(res chan Memory) {
	memInfoContent, err := reader.ReadMemInfo()

	if err != nil {
		panic(err)
	}

	memInfoLines := strings.Split(string(memInfoContent), "\n")

	memInfoMap := make(map[string]int)

	for _, line := range memInfoLines {
		seperatedLine := strings.Split(line, ":")

		if len(seperatedLine) != 2 {
			continue
		}

		key := seperatedLine[0]
		valStr := strings.Fields(seperatedLine[1])[0]

		value, err := strconv.Atoi(valStr)

		if err != nil {
			panic(err)
		}

		memInfoMap[key] = value
	}

	total := memInfoMap["MemTotal"]
	available := memInfoMap["MemAvailable"]
	allocated := total - available
	usage := (float32(allocated) / float32(total)) * 100

	var swap *Memory = nil

	swapTotal := memInfoMap["SwapTotal"]
	if swapTotal != 0 {
		swapAvailable := memInfoMap["SwapFree"]
		swapAllocated := swapTotal - swapAvailable
		swapUsage := (float32(swapAllocated) / float32(swapTotal)) * 100

		swap = &Memory{
			Usage:     swapUsage,
			Total:     swapTotal,
			Available: swapAvailable,
			Allocated: swapAllocated,
		}
	}

	res <- Memory{
		Usage:     usage,
		Total:     total,
		Available: available,
		Allocated: allocated,
		Swap:      swap,
	}
}