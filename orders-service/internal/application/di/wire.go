//go:build wireinject
// +build wireinject

package di

import (
	"orders-service/internal/application/service"
	"orders-service/internal/infrastructure/config"
	"orders-service/internal/infrastructure/persistence/postgres"
	"orders-service/internal/interfaces/api/router"
	"orders-service/internal/interfaces/repository"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	postgres.NewOrdersRepository,
	wire.Bind(new(repository.OrdersRepository), new(*postgres.OrdersRepository)),
)

func InitializeApplication() (*Application, error) {
	wire.Build(
		NewConfigApp,
		config.MustLoad,
		NewPostgresConfig,
		RepositorySet,
		service.NewOrdersService,
		router.NewRouter,
		postgres.NewDb,
		NewApplication,
	)

	return &Application{}, nil
}

func NewConfigApp() *config.App {
	return config.NewApp("config/config.yaml")
}

func NewPostgresConfig(config *config.Config) *postgres.Config {
	return &postgres.Config{
		Host: config.Db.Host,
		Port: config.Db.Port,
		User: config.Db.User,
		Pass: config.Db.Pass,
		Name: config.Db.Name,
	}
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
