package system

import "syscall"

func GetUptime() (int64, error) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)

	if err != nil {
		return 0, err
	}

	return info.Uptime, nil
}
