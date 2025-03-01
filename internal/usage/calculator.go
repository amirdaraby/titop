package usage

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amirdaraby/titop/internal/reader"
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

}

/*
* reads overall usage from /proc/
 */
func Overall(cpuRes chan CPU, memRes chan Memory) {
	go cpuOverallUsage(cpuRes)
	go memOverallUsage(memRes)
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

	res <- Memory{
		Usage:     float32(usage),
		Total:     total,
		Available: available,
		Allocated: allocated,
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
