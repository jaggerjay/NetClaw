package proxy

type Config struct {
	ListenAddress string
	CaptureBodies bool
	MaxBodyBytes  int64
	DataDir       string
}

func DefaultConfig() Config {
	return Config{
		ListenAddress: "127.0.0.1:9090",
		CaptureBodies: true,
		MaxBodyBytes:  1024 * 1024,
		DataDir:       ".netclaw-data",
	}
}
