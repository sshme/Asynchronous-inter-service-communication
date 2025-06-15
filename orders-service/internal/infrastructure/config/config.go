package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Port int `yaml:"port"`
}

type Db struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	Name string `yaml:"name"`
}

type KafkaPublisher struct {
	IntervalMs int `yaml:"interval_ms"`
	BatchSize  int `yaml:"batch_size"`
	MaxRetries int `yaml:"max_retries"`
}

type KafkaConsumer struct {
	GroupID string `yaml:"group_id"`
}

type Kafka struct {
	Publisher KafkaPublisher `yaml:"publisher"`
	Consumer  KafkaConsumer  `yaml:"consumer"`
	Brokers   []string       `yaml:"brokers"`
}

type Redis struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Channel string `yaml:"channel"`
}

type Config struct {
	Server Server `yaml:"server"`
	Db     Db     `yaml:"db"`
	Kafka  Kafka  `yaml:"kafka"`
	Redis  Redis  `yaml:"redis"`
}

func (c *Config) GetPublisherInterval() time.Duration {
	if c.Kafka.Publisher.IntervalMs <= 0 {
		return time.Second
	}
	return time.Duration(c.Kafka.Publisher.IntervalMs) * time.Millisecond
}

func (c *Config) GetPublisherBatchSize() int {
	if c.Kafka.Publisher.BatchSize <= 0 {
		return 50
	}
	return c.Kafka.Publisher.BatchSize
}

func (c *Config) GetPublisherMaxRetries() int {
	if c.Kafka.Publisher.MaxRetries <= 0 {
		return 3
	}
	return c.Kafka.Publisher.MaxRetries
}

type App struct {
	path string
}

func NewApp(path string) *App {
	return &App{
		path: path,
	}
}

func MustLoad(app *App) *Config {
	file, err := os.Open(app.path)
	if err != nil {
		panic(fmt.Sprintf("failed to open config file at %s: %v", app.path, err))
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(fmt.Sprintf("failed to close config file at %s: %v", app.path, err))
		}
	}(file)

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		panic(fmt.Sprintf("failed to decode config file at %s: %v", app.path, err))
	}

	return &config
}

func (k *Kafka) GetBrokers() []string {
	return k.Brokers
}

func (k *Kafka) GetPaymentsEventsTopic() string {
	return "payments.events"
}
