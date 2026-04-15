package proxy

import "time"

type Config struct {
	ListenAddress      string
	CaptureBodies      bool
	MaxBodyBytes       int64
	DataDir            string
	MITMBypassHosts    []string
	MITMFailureBackoff time.Duration
}

func DefaultConfig() Config {
	return Config{
		ListenAddress:      "127.0.0.1:9090",
		CaptureBodies:      true,
		MaxBodyBytes:       1024 * 1024,
		DataDir:            ".netclaw-data",
		MITMBypassHosts:    []string{"localhost", "127.0.0.1"},
		MITMFailureBackoff: 10 * time.Minute,
	}
}
