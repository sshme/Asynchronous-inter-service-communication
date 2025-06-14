//go:build wireinject
// +build wireinject

package di

import (
	"orders-service/internal/application/service"
	"orders-service/internal/infrastructure/brokers/kafka"
	"orders-service/internal/infrastructure/config"
	"orders-service/internal/infrastructure/persistence/postgres"
	"orders-service/internal/infrastructure/sse"
	"orders-service/internal/interfaces/api/router"
	"orders-service/internal/interfaces/repository"
	"orders-service/pkg/random"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	postgres.NewOrdersRepository,
	postgres.NewOutboxRepository,
	postgres.NewInboxRepository,
)

var RandomSet = wire.NewSet(
	random.NewCryptoGenerator,
	wire.Bind(new(random.Generator), new(*random.CryptoGenerator)),
)

var KafkaSet = wire.NewSet(
	kafka.NewConfig,
	NewOutboxPublisher,
	NewInboxProcessor,
)

var SSESet = wire.NewSet(
	sse.NewManager,
)

func InitializeApplication() (*Application, error) {
	wire.Build(
		NewConfigApp,
		config.MustLoad,
		NewPostgresConfig,
		RepositorySet,
		RandomSet,
		KafkaSet,
		SSESet,
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

func NewOutboxPublisher(
	outboxRepo repository.OutboxRepository,
	kafkaConfig *kafka.Config,
) *kafka.OutboxPublisher {
	publisher, err := kafka.NewOutboxPublisher(outboxRepo, kafkaConfig)
	if err != nil {
		panic(err)
	}
	return publisher
}

func NewInboxProcessor(
	inboxRepo repository.InboxRepository,
	kafkaConfig *kafka.Config,
) *kafka.InboxProcessor {
	processor, err := kafka.NewInboxProcessor(inboxRepo, kafkaConfig)
	if err != nil {
		panic(err)
	}
	return processor
}

type Application struct {
	Router          *router.Router
	Config          *config.Config
	OutboxPublisher *kafka.OutboxPublisher
	InboxProcessor  *kafka.InboxProcessor
	OrdersService   *service.OrdersService
	SSEManager      *sse.Manager
}

func NewApplication(
	router *router.Router,
	config *config.Config,
	outboxPublisher *kafka.OutboxPublisher,
	inboxProcessor *kafka.InboxProcessor,
	ordersService *service.OrdersService,
	sseManager *sse.Manager,
) *Application {
	return &Application{
		Router:          router,
		Config:          config,
		OutboxPublisher: outboxPublisher,
		InboxProcessor:  inboxProcessor,
		OrdersService:   ordersService,
		SSEManager:      sseManager,
	}
}
