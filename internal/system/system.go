package system

import "golang.org/x/sys/unix"

func GetUptime() (int64, error) {
	var info unix.Sysinfo_t
	err := unix.Sysinfo(&info)

	if err != nil {
		return 0, err
	}

	return info.Uptime, nil
}
