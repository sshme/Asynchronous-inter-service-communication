//go:build wireinject
// +build wireinject

package di

import (
	"orders-service/internal/infrastructure/config"
	"orders-service/internal/interfaces/api/router"

	"github.com/google/wire"
)

func InitializeApplication() (*Application, error) {
	wire.Build(
		NewConfigApp,
		config.MustLoad,
		router.NewRouter,
		NewApplication,
	)

	return &Application{}, nil
}

func NewConfigApp() *config.App {
	return config.NewApp("config/config.yaml")
}

type Application struct {
	Router *router.Router
	Config *config.Config
}

func NewApplication(router *router.Router, config *config.Config) *Application {
	return &Application{
		Router: router,
		Config: config,
	}
}
