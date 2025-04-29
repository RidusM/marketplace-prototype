package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

//TODO: env default values change

type Config struct {
	GrpcConfig     *gRPCconfig     `yaml:"grpcConfig"`
	PostgresConfig *PostgresConfig `yaml:"postgresConfig"`
	RedisConfig    *RedisConfig    `yaml:"redisConfig"`
	KafkaConfig    KafkaConfig     `yaml:"kafkaConfig"`
	OTLPConfig     OTLPConfig      `yaml:"otlp" env-prefix:"OTLP_"`
	Env            string          `yaml:"env" env:"ENV" env-default:"local"`
}

type gRPCconfig struct {
	Port string `yaml:"port" env:"GRPC_PORT" env-default:""`
}

type PostgresConfig struct {
	User         string        `yaml:"user" env:"POSTGRES_USER" env-default:""`
	Password     string        `yaml:"password" env:"POSTGRES_PASSWORD" env-default:""`
	Host         string        `yaml:"host" env:"POSTGRES_HOST" env-default:""`
	Port         string        `yaml:"port" env:"POSTGRES_PORT" env-default:""`
	Name         string        `yaml:"name" env:"POSTGRES_NAME" env-default:""`
	Connection   string        `yaml:"connection" env:"POSTGRES_CONNECTION" env-default:""`
	MaxPool      int           `yaml:"maxPool" env:"POSTGRES_MAX_POOL" env-default:""`
	ConnTimeout  time.Duration `yaml:"timeout" env:"POSTGRES_CONN_TIMEOUT" env-default:""`
	ConnAttempts int           `yaml:"connAttempts" env:"POSTGRES_CONN_ATTEMPTS" env-default:""`
}

type RedisConfig struct {
	Password    string        `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	Addr        string        `yaml:"address" env:"REDIS_ADDRESS" env-default:""`
	PoolSize    int           `yaml:"poolSize" env:"REDIS_POOL_SIZE" env-default:""`
	MinIdleCons int           `yaml:"minIdleCons" env:"REDIS_MIN_IDLE_CONS" env-default:""`
	ConnTimeout time.Duration `yaml:"connTimeout" env:"REDIS_CONN_TIMEOUT" env-default:""`
}

type KafkaConfig struct {
	Brokers []string `env:"BROKERS,required" yaml:"brokers"`
}

type OTLPConfig struct {
	OTLPEndpoint string `env:"ENDPOINT,required" yaml:"endpoint"`
}

func New(configPath string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
