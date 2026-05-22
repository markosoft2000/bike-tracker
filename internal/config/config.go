package config

import (
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
}

type HTTPServerConfig struct {
	Address string `yaml:"address" env:"HTTP_PORT" env-default:"localhost:8080"`
	Timeout int    `yaml:"timeout" env-default:"3"`
}

var (
	cfg  *Config
	once sync.Once
)

func init() {
	// Register the flag in the global set so that 'go test' and other tools
	// using flag.Parse() recognize it and don't fail during initialization.
	if flag.Lookup("config") == nil {
		flag.String("config", "", "path to config file")
	}
}

func MustLoad() *Config {
	once.Do(func() {
		var configPath string

		// Manually look for -config flag to avoid conflicts with other FlagSets
		// used in tools like migrators or test runners.
		for i := 0; i < len(os.Args); i++ {
			arg := os.Args[i]
			if strings.HasPrefix(arg, "-config=") || strings.HasPrefix(arg, "--config=") {
				parts := strings.SplitN(arg, "=", 2)
				configPath = parts[1]
				break
			} else if arg == "-config" || arg == "--config" {
				if i+1 < len(os.Args) {
					configPath = os.Args[i+1]
					break
				}
			}
		}

		if configPath == "" {
			configPath = os.Getenv("CONFIG_PATH")
		}

		if configPath == "" {
			panic("CONFIG_PATH is not set")
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			panic("config file does not exist: " + configPath)
		}

		cfg = &Config{}
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			panic("cannot read config: " + err.Error())
		}
	})

	return cfg
}
