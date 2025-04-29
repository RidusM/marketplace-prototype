package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App      App            `yaml:"app" env-prefix:"APP_"`
		Auth     Auth           `yaml:"auth" env-prefix:"AUTH_"`
		Postgres PostgresConfig `yaml:"postgres" env-prefix:"DB_"`
		Redis    RedisConfig    `yaml:"redis" env-prefix:"REDIS_"`
		Kafka    KafkaConfig    `yaml:"kafka" env-prefix:"KAFKA_"`
		OTLP     OTLPConfig     `yaml:"otlp" env-prefix:"OTLP_"`
		Email    EmailConfig    `yaml:"email" env-prefix:"EMAIL_"`
		Metrics  MetricsConfig  `yaml:"metrics" env-prefix:"METRICS_"`
		Env      string         `yaml:"env" env:"ENV" env-default:"local"`
	}

	App struct {
		Port    string `env:"PORT,required" yaml:"port"`
		Name    string `env:"NAME,required" yaml:"name"`
		Version string `env:"VERSION,required" yaml:"version"`
	}

	Auth struct {
		AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL" yaml:"access_token_ttl"`
		RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL" yaml:"refresh_token_ttl"`
		VerifyTTL       time.Duration `env:"VERIFY_TTL" yaml:"verify_ttl"`
		UserCacheTTL    time.Duration `env:"USER_CACHE_TTL" yaml:"user_cache_ttl"`
		SecretKey       string        `env:"SECRET,required" yaml:"secret_key"`
	}

	PostgresConfig struct {
		Connection string `env:"CONNECTION,required" yaml:"connection"`
		Host       string `env:"HOST,required" yaml:"host"`
		Port       string `env:"PORT,required" yaml:"port"`
		Name       string `env:"NAME,required" yaml:"name"`
		User       string `env:"USER,required" yaml:"user"`
		Password   string `env:"PASSWORD,required" yaml:"password"`
		PoolMax    int    `env:"POOL_MAX,required" yaml:"pool_max"`
	}

	RedisConfig struct {
		Addr     string `env:"ADDR,required" yaml:"addr"`
		Password string `env:"PASSWORD,required" yaml:"password"`
	}

	KafkaConfig struct {
		Brokers []string `env:"BROKERS,required" yaml:"brokers"`
	}

	OTLPConfig struct {
		OTLPEndpoint string `env:"ENDPOINT,required" yaml:"endpoint"`
	}

	MetricsConfig struct {
		MetricsPort string `env:"PORT,required" yaml:"port"`
	}

	EmailConfig struct {
		Host     string `env:"HOST,required" yaml:"host"`
		Port     int    `env:"PORT,required" yaml:"port"`
		Username string `env:"USERNAME,required" yaml:"username"`
		Password string `env:"PASSWORD,required" yaml:"password"`
		Sender   string `env:"SENDER,required" yaml:"sender"`
	}
)

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
