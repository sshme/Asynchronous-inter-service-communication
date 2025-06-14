package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Db struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		Name string `yaml:"name"`
	} `yaml:"db"`
	Kafka struct {
		Publisher struct {
			IntervalMs int `yaml:"interval_ms"`
			BatchSize  int `yaml:"batch_size"`
			MaxRetries int `yaml:"max_retries"`
		} `yaml:"publisher"`
		Brokers []string `yaml:"brokers"`
	} `yaml:"kafka"`
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
