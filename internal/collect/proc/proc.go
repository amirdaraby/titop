package proc

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/amirdaraby/titop/internal/reader"
	"github.com/amirdaraby/titop/internal/shared"
)

var processLastStates map[string]processStat = make(map[string]processStat)

type Process struct {
	ID       string
	Command  string
	State    string
	Priority string
	CpuUsage float32
	MemUsage float32
	IO       int64 // bytes
}

type processStat struct {
	uTime, sTime, startTime, systemUptime, readBytes, writeBytes int64
}

const (
	ID_PROCESS = iota
	COMM_PROCESS
	STATE_PROCESS
	PARENT_ID_PROCESS
	GROUP_ID_PROCESS
	SESSION_ID_PROCESS
	TTY_NR_PROCESS
	TPG_ID_PROCESS
	FLAGS_PROCESS
	MINFLT_PROCESS
	CMINFLT_PROCESS
	MAJFLT_PROCESS
	CMAJFLT_PROCESS
	UTIME_PROCESS
	STIME_PROCESS
	CUTIME_PROCESS
	CSTIME_PROCESS
	PRIORITY_PROCESS
	NICE_PROCESS
	NUM_THREADS_PROCESS
	IT_REAL_VALUE_PROCESS
	START_TIME_PROCESS
	VSIZE_PROCESS
	RSS_PROCESS
	RSSLIM_PROCESS
	START_CODE_PROCESS
	END_CODE_PROCESS
	START_STACK_PROCESS
	KST_KESP_PROCESS
	KST_KEIP_PROCESS
	SIGNAL_PROCESS
	BLOCKED_PROCESS
	SIG_IGNORE_PROCESS
	SIG_CATCH_PROCESS
	WCHAN_PROCESS
	NSWAP_PROCESS
	CNSWAP_PROCESS
	EXIT_SIGNAL_PROCESS
	PROCESSOR_PROCESS
	RT_PRIORITY_PROCESS
	POLICY_PROCESS
	BLK_IO_TICKS_PROCESS
	GTIME_PROCESS
	CGTIME_PROCESS
	START_DATA_PROCESS
	END_DATA_PROCESS
	START_BRK_PROCESS
	ARG_START_PROCESS
	ARG_END_PROCESS
	ENV_START_PROCESS
	ENV_END_PROCESS
	EXIT_CODE_PROCESS
)

const (
	M_VSIZE_PROCESS = iota
	M_RSS_PROCESS
	M_SWAP_SHARED_PROCESS
	M_SWAP_TEXT_PROCESS
	M_SWAP_LIBRARY_PROCESS
	M_SWAP_DATA_PROCESS
	M_SWAP_DIRTY_PROCESS
)

func (p *processStat) processTime() int64 {
	return p.uTime + p.sTime
}

func SendUsage(res chan []Process) {
	processesContent := reader.ReadProcesses()

	var processes []Process
	seenPIDs := make(map[string]struct{})

	systemUptime, err := shared.GetUptime()
	if err != nil {
		panic(err)
	}

	numCores := shared.GetConfig().CoresCount

	for _, p := range processesContent {
		stats := strings.Split(string(p["stat"]), " ")

		pid := stats[ID_PROCESS]
		if _, exists := seenPIDs[pid]; exists {
			continue
		}

		seenPIDs[pid] = struct{}{}

		cmd := stats[COMM_PROCESS]
		priority := stats[PRIORITY_PROCESS]
		state := stats[STATE_PROCESS]

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

		mstats := strings.Split(string(p["statm"]), " ")

		rss, err := strconv.Atoi(mstats[M_RSS_PROCESS])

		if err != nil {
			panic(err)
		}

		memAlloc := int64(rss) * shared.GetConfig().PageSize

		memUsage := float32(memAlloc) / float32(shared.GetConfig().TotalMem) * 100

		ioContent, ioExists := p["io"]

		var ioBytes int64 = -1
		var readBytes int64 = 0
		var writeBytes int64 = 0

		if ioExists {
			ioStatLines := strings.Split(string(ioContent), "\n")
			ioStatMap := make(map[string]int64)

			for _, line := range ioStatLines {
				seperatedLine := strings.Split(line, ":")

				if len(seperatedLine) != 2 {
					continue
				}

				key := seperatedLine[0]
				valStr := strings.Fields(seperatedLine[1])[0]

				value, err := strconv.ParseInt(valStr, 10, 64)

				if err != nil {
					panic(err)
				}

				ioStatMap[key] = value
			}

			readBytes = ioStatMap["read_bytes"]
			writeBytes = ioStatMap["write_bytes"]
			ioBytes = 0
		}

		currentStat := processStat{
			uTime:        int64(utime),
			sTime:        int64(stime),
			startTime:    int64(startTime),
			systemUptime: systemUptime,
			readBytes:    readBytes,
			writeBytes:   writeBytes,
		}

		lastStat, exists := processLastStates[pid]
		if !exists {
			lastStat = currentStat
			processLastStates[pid] = currentStat
			processes = append(processes, Process{
				ID:       pid,
				Command:  cmd,
				State:    state,
				Priority: priority,
				CpuUsage: 0,
				MemUsage: memUsage,
				IO:       ioBytes,
			})
			continue
		}

		processTimeDiff := float64(currentStat.processTime() - lastStat.processTime())

		uptimeDiff := float64(systemUptime-lastStat.systemUptime) * float64(shared.GetConfig().ClkTck)

		var cpuUsage float32
		if uptimeDiff > 0 {
			cpuUsage = float32((processTimeDiff / uptimeDiff) * 100.0 / float64(numCores))
		}

		if cpuUsage < 0 || math.IsNaN(float64(cpuUsage)) {
			cpuUsage = 0
		} else if cpuUsage > 100.0 {
			cpuUsage = 100.0
		}

		if ioBytes != -1 {
			readD := readBytes - lastStat.readBytes
			writeD := writeBytes - lastStat.writeBytes

			elapsedTime := time.Now().Sub(shared.GetLastRefresh()).Seconds()

			if elapsedTime >= 1 {
				ioBytes = (readD + writeD) / int64(elapsedTime)
			}
		}

		processLastStates[pid] = currentStat

		processes = append(processes, Process{
			ID:       pid,
			Command:  cmd,
			State:    state,
			Priority: priority,
			CpuUsage: cpuUsage,
			MemUsage: memUsage,
			IO:       ioBytes,
		})
	}

	res <- processes
}
