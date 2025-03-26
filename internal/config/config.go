package config

import (
	"github.com/tklauser/go-sysconf"
)

type Config struct {
	System
}

var config *Config

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

	system := System{
		ClkTck:     clktck,
		CoresCount: numCores,
		PageSize:   pageSize,
		TotalMem:   totalMem,
	}

	config = &Config{
		System: system,
	}

	return nil
}

func Get() *Config {

	if config == nil {
		panic("config is not initalized")
	}

	return config
}
