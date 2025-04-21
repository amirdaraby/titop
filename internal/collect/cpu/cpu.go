package cpu

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amirdaraby/titop/internal/reader"
)

var overallCpuLastStats []cpuCoreOverallStat

type Core struct {
	Usage float32
}

type cpuCoreOverallStat struct {
	stat [10]int
}

type CPU struct {
	Usage  float32
	UpTime time.Duration
	Cores  []Core
}

type cpuStat struct {
	userOverallStat, niceOverallStat, systemOverallStat, idleOverallStat, ioWaitOverallStat, irqOverallStat, softIrqOverallStat, stealOverallStat int
}

const (
	USER_OVERALL_STAT = iota
	NICE_OVERALL_STAT
	SYSTEM_OVERALL_STAT
	IDLE_OVERALL_STAT
	IOWAIT_OVERALL_STAT
	IRQ_OVERALL_STAT
	SOFTIRQ_OVERALL_STAT
	STEAL_OVERALL_STAT
	GUEST_OVERALL_STAT
	GUEST_NICE_OVERALL_STAT
)


func SendUsage(res chan CPU) {

	cpuStatContent, err := reader.ReadStat()

	if err != nil {
		panic(err)
	}

	uptimeContent, err := reader.ReadUptime()

	if err != nil {
		panic(err)
	}

	uptimeInSeconds := strings.Split(string(uptimeContent), " ")[0]

	regex, err := regexp.Compile(`(?m)^cpu\d+.*$`)

	if err != nil {
		panic(err)
	}

	coreLines := regex.FindAllString(string(cpuStatContent), -1)

	var currentCoreStatuses []cpuCoreOverallStat

	for _, c := range coreLines {
		spiltedData := strings.Split(c, " ")

		var coreStats [10]int
		for i := 1; i < len(spiltedData); i++ {
			coreStats[i-1], err = strconv.Atoi(spiltedData[i])
			if err != nil {
				panic(err)
			}
		}

		currentCoreStatuses = append(currentCoreStatuses, cpuCoreOverallStat{
			stat: coreStats,
		})
	}

	if err != nil {
		panic(err)
	}

	cpu := calculateCpuCoresOverallUsage(currentCoreStatuses)
	cpu.UpTime, err = time.ParseDuration(fmt.Sprintf("%s%s", uptimeInSeconds, "s"))

	if err != nil {
		panic(err)
	}

	res <- cpu
}

func calculateCpuCoresOverallUsage(coreStats []cpuCoreOverallStat) CPU {
	cpu := CPU{}

	var totalUsage float32

	if len(coreStats) != len(overallCpuLastStats) {
		overallCpuLastStats = make([]cpuCoreOverallStat, len(coreStats))
		copy(overallCpuLastStats, coreStats)
	}

	for key, currentStat := range coreStats {
		lastStat := overallCpuLastStats[key]

		currentTotalTime := overallCpuTotalTime(currentStat.stat)
		currentIdleTime := overallCpuIdleTime(currentStat.stat)

		lastTotalTime := overallCpuTotalTime(lastStat.stat)
		lastIdleTime := overallCpuIdleTime(lastStat.stat)

		totalDelta := currentTotalTime - lastTotalTime
		idleDelta := currentIdleTime - lastIdleTime

		usage := float32(100 * (float32((totalDelta - idleDelta)) / float32(totalDelta)))

		if math.IsNaN(float64(usage)) {
			usage = 0
		}

		core := Core{
			Usage: usage,
		}

		cpu.Cores = append(cpu.Cores, core)
		totalUsage += core.Usage

		overallCpuLastStats[key] = currentStat
	}

	cpu.Usage = totalUsage / float32(len(cpu.Cores))

	return cpu
}

func overallCpuTotalTime(stat [10]int) int {
	return stat[USER_OVERALL_STAT] + stat[NICE_OVERALL_STAT] + stat[SYSTEM_OVERALL_STAT] + stat[IDLE_OVERALL_STAT] + stat[IOWAIT_OVERALL_STAT] + stat[IRQ_OVERALL_STAT] + stat[SOFTIRQ_OVERALL_STAT] + stat[STEAL_OVERALL_STAT]
}

func overallCpuIdleTime(stat [10]int) int {
	return stat[IDLE_OVERALL_STAT] + stat[IOWAIT_OVERALL_STAT]
}
