package reader

import "os"

func ReadMemInfo() (memInfoContent []byte, err error) {

	memInfoContent, err = os.ReadFile("/proc/meminfo")

	return
}

func ReadStat() (statContent []byte, err error) {

	statContent, err = os.ReadFile("/proc/stat")

	return
}

func ReadUptime() (uptimeContent []byte, err error) {
	uptimeContent, err = os.ReadFile("/proc/uptime")

	return
}