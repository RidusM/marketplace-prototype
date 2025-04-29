package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type PostgresConfig struct {
	Connection   string        `yaml:"connection"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	Name         string        `yaml:"name"`
	MaxPool      int           `yaml:"maxPool" env:"POSTGRES_MAX_POOL" env-default:""`
	ConnTimeout  time.Duration `yaml:"timeout" env:"POSTGRES_CONN_TIMEOUT" env-default:""`
	ConnAttempts int           `yaml:"connAttempts" env:"POSTGRES_CONN_ATTEMPTS" env-default:""`
}

type RedisConfig struct {
	Password    string        `yaml:"password"`
	Addr        string        `yaml:"addr"`
	PoolSize    int           `yaml:"poolSize" env:"REDIS_POOL_SIZE" env-default:""`
	MinIdleCons int           `yaml:"minIdleCons" env:"REDIS_MIN_IDLE_CONS" env-default:""`
	ConnTimeout time.Duration `yaml:"connTimeout" env:"REDIS_CONN_TIMEOUT" env-default:""`
}

type OTLPConfig struct {
	OTLPEndpoint string `env:"ENDPOINT,required" yaml:"endpoint"`
}

type KafkaConfig struct {
	Brokers []string `env:"BROKERS,required" yaml:"brokers"`
}

type Config struct {
	Env      string          `yaml:"env"`
	GRPCPort int             `yaml:"grpc_port"`
	OTLP     OTLPConfig      `yaml:"otlp" env-prefix:"OTLP_"`
	Postgres *PostgresConfig `yaml:"postgres"`
	Redis    *RedisConfig    `yaml:"redis"`
	Kafka    *KafkaConfig    `yaml:"kafka"`
}

func New(configPath string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
