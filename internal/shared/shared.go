package shared

import (
	"github.com/tklauser/go-sysconf"
	"golang.org/x/sys/unix"
)

type Config struct {
	ClkTck int64
	CoresCount int64
	PageSize int64
	TotalMem int64 // in RSS
}

var cfg *Config

func Init() error {
	clktck, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)

	if err != nil {
		return err
	}

	numCores, err := sysconf.Sysconf(sysconf.SC_NPROCESSORS_ONLN)
	if err != nil {
		return err
	}

	pageSize, err := sysconf.Sysconf(sysconf.SC_PAGESIZE)

	if err != nil {
		return err
	}

	memPages, err := sysconf.Sysconf(sysconf.SC_PHYS_PAGES)

	if err != nil {
		return err
	}

	totalMem := memPages * pageSize

	cfg = &Config{
		ClkTck:     clktck,
		CoresCount: numCores,
		PageSize:   pageSize,
		TotalMem:   totalMem,
	}

	return nil
}

func GetUptime() (int64, error) {
	var info unix.Sysinfo_t
	err := unix.Sysinfo(&info)

	if err != nil {
		return 0, err
	}

	return info.Uptime, nil
}

func GetConfig() *Config {
	if cfg == nil {
		panic("Initialize Shared before using this function")
	}

	return cfg
}