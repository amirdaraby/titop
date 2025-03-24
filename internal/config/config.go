package config

import "github.com/tklauser/go-sysconf"

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

	system := System{
		ClkTck: clktck,
		CoresCount:  numCores,
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
