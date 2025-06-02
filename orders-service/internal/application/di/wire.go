//go:build wireinject
// +build wireinject

package di

import (
	"orders-service/internal/infrastructure/config"

	"github.com/google/wire"
)

func InitializeApplication() (*Application, error) {
	wire.Build(
		NewConfigApp,
		config.MustLoad,

		NewApplication,
	)

	return &Application{}, nil
}

func NewConfigApp() *config.App {
	return config.NewApp("config/config.yaml")
}

type Application struct {
	Config *config.Config
}

func NewApplication(config *config.Config) *Application {
	return &Application{
		Config: config,
	}
}
