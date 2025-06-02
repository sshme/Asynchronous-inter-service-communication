package config

import (
	"fmt"
	"os"

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
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		panic(fmt.Sprintf("failed to decode config file at %s: %v", app.path, err))
	}

	return &config
}
