package reader

import (
	"fmt"
	"os"
	"path/filepath"
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
 		filepath.Walk("/proc", func(path string, info os.FileInfo, err error) error {

			if err != nil || !info.IsDir() {
				return nil
			}
	
			name := info.Name()
	
			_, err = strconv.Atoi(name)
	
			if err != nil {
				return nil
			}
	
			statContent, err := os.ReadFile(fmt.Sprintf("/proc/%s/stat", name))
	
			if err != nil {
				return nil
			}
	
			memStatContent, err := os.ReadFile(fmt.Sprintf("/proc/%s/statm", name))
	
			if err != nil {
				return nil
			}
	
			processMap := make(map[string][]byte)

			processMap["stat"] = statContent
			processMap["statm"] = memStatContent

			processesContent = append(processesContent, processMap)
	
			return nil
		})

	return processesContent
}
