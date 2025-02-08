package processes

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

const CLK_TK = 100 // todo check this from system config in compile time

const (
	USER = iota
	NICE
	SYSTEM
	IDLE
	IOWAIT
	IRQ
	SOFTIRQ
	STEAL
	GUEST
	GUESTNICE
)

func CpuOverall() {

	T1Data := getCpuStatFromProcStatFile()

	for {
		T1TotalTime := getCpuTotalTime(T1Data)
		T1IdleTime := getCpuIdleTime(T1Data)

		time.Sleep(time.Second * 2)

		T2Data := getCpuStatFromProcStatFile()
		T2TotalTime := getCpuTotalTime(T2Data)
		T2IdleTime := getCpuIdleTime(T2Data)

		TotalDelta := T2TotalTime - T1TotalTime
		IdleDelta := T2IdleTime - T1IdleTime

		CpuUsage := 100 * (float64((TotalDelta - IdleDelta)) / float64(TotalDelta))

		clear := exec.Command("clear")

		clear.Stdout = os.Stdout

		clear.Run()

		fmt.Printf("%.2f%%", CpuUsage)

		T1Data = T2Data
	}

}

func getCpuTotalTime(stat [10]int) int {
	return stat[USER] + stat[NICE] + stat[SYSTEM] + stat[IDLE] + stat[IOWAIT] + stat[IRQ] + stat[SOFTIRQ] + stat[STEAL]
}

func getCpuIdleTime(stat [10]int) int {
	return stat[IDLE] + stat[IOWAIT]
}

func getCpuStatFromProcStatFile() [10]int {
	content, err := os.ReadFile("/proc/stat")

	if err != nil {
		panic(err)
	}

	r := regexp.MustCompile("^(.*)")

	cpuOverall := r.Find(content)

	r = regexp.MustCompile("[0-9]+")

	stat := r.FindAllString(string(cpuOverall), 10)

	var overallStat [10]int

	for k, s := range stat {
		overallStat[k], err = strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
	}

	return overallStat
}
