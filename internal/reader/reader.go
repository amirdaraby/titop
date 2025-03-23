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

func ReadProcesses() (processesContent [][]byte) {
	processesContentChan := make(chan []byte, 1)
	
	go func () {
		filepath.Walk("/proc", func(path string, info os.FileInfo, err error) error {

			if err != nil || !info.IsDir() {
				return nil
			}
	
			name := info.Name()
	
			_, err = strconv.Atoi(name)
	
			if err != nil {
				return nil
			}
	
			pContent, err := os.ReadFile(fmt.Sprintf("/proc/%s/stat", name))
	
			if err != nil {
				return nil
			}
	
			processesContentChan <- pContent
	
			return nil
		})
		close(processesContentChan)
	}()

	for pContent := range processesContentChan {
		processesContent = append(processesContent, pContent)
	}

	return
}
