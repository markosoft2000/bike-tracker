package config

import (
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env" env:"ENV" env-default:"local"`
	TokenTTL   time.Duration    `yaml:"token_ttl" env:"TOKEN_TTL" env-default:"15m"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Services   ServicesConfig   `yaml:"services"`
	Redis      RedisConfig      `yaml:"redis"`
	Kafka      KafkaConfig      `yaml:"kafka"`
}

type HTTPServerConfig struct {
	Address            string `yaml:"address" env:"HTTP_PORT" env-default:"localhost:8080"`
	ServerHeader       string `yaml:"server_header" env:"HTTP_SERVER_HEADER" env-default:"bike-tracker"`
	DisableKeepalive   bool   `yaml:"disable_keepalive" env:"HTTP_DISABLE_KEEPALIVE" env-default:"false"`
	Concurrency        int    `yaml:"concurrency" env:"HTTP_CONCURRENCY" env-default:"262144"`                // Maximum concurrent connections allocated
	ReduceMemoryUsage  bool   `yaml:"reduce_memory_usage" env:"HTTP_REDUCE_MEMORY_USAGE" env-default:"false"` // CRITICAL: Leaves buffers in pool. Setting true hurts CPU.
	DisableDefaultDate bool   `yaml:"disable_default_date" env:"HTTP_DISABLE_DEFAULT_DATE" env-default:"false"`

	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"10s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"2s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"2s"`
}

type MiddlewareConfig struct {
	RateLimitMax        int           `yaml:"rate_limit_max" env:"RATE_LIMIT_MAX" env-default:"1000"`
	RateLimitExpiration time.Duration `yaml:"rate_limit_expiration" env:"RATE_LIMIT_EXPIRATION" env-default:"60s"`
	TokenIssuer         string        `yaml:"token_issuer" env:"TOKEN_ISSUER" env-default:"markosoft2000"`
	TokenAudience       string        `yaml:"token_audience" env:"TOKEN_AUDIENCE" env-default:"auth-service"`
}

type ServicesConfig struct {
	AuthServiceAddr string    `yaml:"auth_service_addr" env:"AUTH_SERVICE_ADDR" env-default:"localhost:50051"`
	SLO             SLOConfig `yaml:"slo"`
}

type SLOConfig struct {
	Auth SLOConfigAuth `yaml:"auth"`
}

type SLOConfigAuth struct {
	UserRegisterTimeout time.Duration `yaml:"user_register_timeout" env:"USER_REGISTER_TIMEOUT" env-default:"2s"`
	UserLoginTimeout    time.Duration `yaml:"user_login_timeout" env:"USER_LOGIN_TIMEOUT" env-default:"2s"`
	UserLogoutTimeout   time.Duration `yaml:"user_logout_timeout" env:"USER_LOGOUT_TIMEOUT" env-default:"2s"`
	UserIsAdminTimeout  time.Duration `yaml:"user_is_admin_timeout" env:"USER_IS_ADMIN_TIMEOUT" env-default:"2s"`
	RefreshTokenTimeout time.Duration `yaml:"refresh_token_timeout" env:"REFRESH_TOKEN_TIMEOUT" env-default:"2s"`
	AppAddTimeout       time.Duration `yaml:"app_add_timeout" env:"APP_ADD_TIMEOUT" env-default:"2s"`
	AppRemoveTimeout    time.Duration `yaml:"app_remove_timeout" env:"APP_REMOVE_TIMEOUT" env-default:"2s"`
}

type RedisConfig struct {
	Host             string        `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port             int           `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	OperationTimeout time.Duration `yaml:"operation_timeout" env:"OPERATION_TIMEOUT" env-default:"5s"`
}

type KafkaConfig struct {
	Brokers                     string `yaml:"brokers" env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	UserActivityGroupID         string `yaml:"user_activity_group_id" env:"KAFKA_USER_ACTIVITY_GROUP_ID" env-default:"gateway-act-group"`
	UserActivityTopic           string `yaml:"user_activity_topic" env:"KAFKA_USER_ACTIVITY_TOPIC" env-default:"auth-user-activity-v1"`
	UserActivityAutoOffsetReset string `yaml:"user_activity_auto_offset_reset" env:"KAFKA_USER_ACTIVITY_AUTO_OFFSET_RESET" env-default:"earliest"`

	AppPublicKeyGroupID         string `yaml:"app_public_key_group_id" env:"KAFKA_APP_PUBLIC_KEY_GROUP_ID" env-default:"gateway-app-pk-group"`
	AppPublicKeyTopic           string `yaml:"app_public_key_topic" env:"KAFKA_APP_PUBLIC_KEY_TOPIC" env-default:"auth-app-key-v1"`
	AppPublicKeyAutoOffsetReset string `yaml:"app_public_key_auto_offset_reset" env:"KAFKA_APP_PUBLIC_KEY_AUTO_OFFSET_RESET" env-default:"earliest"`
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
