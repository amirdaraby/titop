package usage

import (
	"time"
)

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

type Memory struct {
	Usage                       float32
	Total, Available, Allocated int // KB
	Swap                        *Memory
}

type Process struct {
	ID       int
	Command  string
	State    string
	Priority string
	CpuUsage float32
	MemUsage float32
}

type processCpuStat struct {
	uTime, sTime, startTime int
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
	RS_SLIM_PROCESS
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
	VM_SIZE_MEM = 16
	VM_RSS_MEM  = 20
	VM_SWAP_MEM = 25
)
