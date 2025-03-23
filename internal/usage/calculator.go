package usage

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amirdaraby/titop/internal/config"
	"github.com/amirdaraby/titop/internal/reader"
	"github.com/amirdaraby/titop/internal/system"
)

/*
*  used to remember last state of processes to calculate usage
 */
var processLastStates map[string]processCpuStat = make(map[string]processCpuStat)

var overallCpuLastStats []cpuCoreOverallStat

/*
** read processes usage from /proc/pid directories
 */
func Processes(res chan []Process) {
	processesContent := reader.ReadProcesses()

	var processes []Process

	systemUptime, err := system.GetUptime()

	if err != nil {
		panic(err)
	}

	for _, p := range processesContent {
		stats := strings.Split(string(p), " ")

		cmd := stats[COMM_PROCESS]
		priority := stats[PRIORITY_PROCESS]
		state := stats[STATE_PROCESS]
		pid := stats[ID_PROCESS]

		utime, err := strconv.Atoi(stats[UTIME_PROCESS])
		if err != nil {
			panic(err)
		}

		stime, err := strconv.Atoi(stats[STIME_PROCESS])
		if err != nil {
			panic(err)
		}

		startTime, err := strconv.Atoi(stats[START_TIME_PROCESS])

		if err != nil {
			panic(err)
		}

		currentStat := processCpuStat{
			uTime:        int64(utime),
			sTime:        int64(stime),
			startTime:    int64(startTime),
			systemUptime: systemUptime,
		}

		lastStat, exists := processLastStates[pid]

		if !exists {
			lastStat = currentStat
		}
		processLastStates[pid] = currentStat

		proccessTimeDiff := (currentStat.processTime() - lastStat.processTime()) / config.Get().ClkTck
		systemUptimeDiff := currentStat.systemUptime - lastStat.systemUptime

		cpuUsage := (float32(proccessTimeDiff) / float32(systemUptimeDiff)) * 100

		if math.IsNaN(float64(cpuUsage)) {
			cpuUsage = 0
		}

		processes = append(processes, Process{
			ID:       pid,
			Command:  cmd,
			State:    state,
			Priority: priority,
			CpuUsage: cpuUsage,
		})
	}

	res <- processes
}

func Calc(cpuRes chan CPU, memRes chan Memory, processesRes chan []Process) {
	go cpuOverallUsage(cpuRes)
	go memOverallUsage(memRes)
	go Processes(processesRes)
}

func memOverallUsage(res chan Memory) {
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

func cpuOverallUsage(res chan CPU) {

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
