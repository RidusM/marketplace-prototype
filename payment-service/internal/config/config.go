package config

import "github.com/ilyakaznacheev/cleanenv"

type PostgresConfig struct {
	Connection string `yaml:"connection"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Name       string `yaml:"name"`
}

type RedisConfig struct {
	Password string `yaml:"password"`
	Addr     string `yaml:"addr"`
}

type Config struct {
	GRPCPort int            `yaml:"grpc_port"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka" env-prefix:"KAFKA_"`
	OTLP     OTLPConfig     `yaml:"otlp" env-prefix:"OTLP_"`
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
