package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App            App           `yaml:"app" env-prefix:"APP_"`
		Kafka          KafkaConfig   `yaml:"kafka" env-prefix:"KAFKA_"`
		OTLP           OTLPConfig    `yaml:"otlp" env-prefix:"OTLP_"`
		ProductService ProductConfig `yaml:"product_service" env-prefix:"PRODUCT_SERVICE_"`
		UserService    UserConfig    `yaml:"user_service" env-prefix:"USER_SERVICE_"`
		PaymentService PaymentConfig `yaml:"payment_service" env-prefix:"PAYMENT_SERVICE_"`
		AuthService    AuthConfig    `yaml:"auth_service" env-prefix:"AUTH_SERVICE_"`
		OrderService   OrderConfig   `yaml:"order_service" env-prefix:"ORDER_SERVICE_"`

		Env string `yaml:"env" env:"ENV" env-default:"local"`
	}

	App struct {
		Name    string `env:"NAME,required" yaml:"name"`
		Version string `env:"VERSION,required" yaml:"version"`
	}

	KafkaConfig struct {
		Brokers []string `env:"BROKERS,required" yaml:"brokers"`
	}

	OTLPConfig struct {
		Endpoint string `env:"ENDPOINT,required" yaml:"endpoint"`
	}

	ProductConfig struct {
		Host        string `env:"HOST,required" yaml:"host"`
		MetricsPort string `env:"METRICS_PORT,required" yaml:"metrics_port"`
	}

	UserConfig struct {
		Host        string `env:"HOST,required" yaml:"host"`
		MetricsPort string `env:"METRICS_PORT,required" yaml:"metrics_port"`
	}

	PaymentConfig struct {
		Host        string `env:"HOST,required" yaml:"host"`
		MetricsPort string `env:"METRICS_PORT,required" yaml:"metrics_port"`
	}

	AuthConfig struct {
		Host        string `env:"HOST,required" yaml:"host"`
		MetricsPort string `env:"METRICS_PORT,required" yaml:"metrics_port"`
	}

	OrderConfig struct {
		Host        string `env:"HOST,required" yaml:"host"`
		MetricsPort string `env:"METRICS_PORT,required" yaml:"metrics_port"`
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
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

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
