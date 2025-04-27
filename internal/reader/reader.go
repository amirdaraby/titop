package reader

import (
	"os"
	"strconv"
)

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

func ReadProcesses() (processesContent []map[string][]byte) {
	dirEntries, err := os.ReadDir("/proc")

	if err != nil {
		panic(err)
	}

	for _, d := range dirEntries {
		if !d.IsDir() {
			break
		}

		dirName := d.Name()
		_, err = strconv.Atoi(dirName)

		if err != nil {
			break
		}

		statContent, err := os.ReadFile("/proc/" + dirName + "/stat")

		if err != nil {
			continue
		}

		memStatContent, err := os.ReadFile("/proc/" + dirName + "/statm")

		if err != nil {
			continue
		}

		processMap := make(map[string][]byte)

		diskStatContent, err := os.ReadFile("/proc/"+dirName+"/io")

		if err == nil {
			processMap["io"] = diskStatContent
		}

		processMap["stat"] = statContent
		processMap["statm"] = memStatContent

		processesContent = append(processesContent, processMap)
	}

	return processesContent
}

func ReadDiskStat() (diskStatContent []byte, err error) {
	diskStatContent, err = os.ReadFile("/proc/diskstats")
	
	return
}
