//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"fmt"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"orders-service/internal/application/service"
	"orders-service/internal/infrastructure/brokers/kafka"
	"orders-service/internal/infrastructure/config"
	"orders-service/internal/infrastructure/persistence/postgres"
	redispubsub "orders-service/internal/infrastructure/pubsub/redis"
	"orders-service/internal/infrastructure/sse"
	"orders-service/internal/interfaces/api/handler"
	"orders-service/internal/interfaces/api/router"
	"orders-service/internal/interfaces/repository"
	"orders-service/pkg/random"
)

func InitializeApplication() (*Application, func(), error) {
	wire.Build(
		NewConfigApp,
		config.MustLoad,
		NewPostgresConfig,
		postgres.NewDb,
		postgres.NewOrdersRepository,
		postgres.NewOutboxRepository,
		postgres.NewInboxRepository,
		random.NewCryptoGenerator,
		wire.Bind(new(random.Generator), new(*random.CryptoGenerator)),
		kafka.NewConfig,
		NewOutboxPublisher,
		NewInboxProcessor,
		NewRedisConfig,
		NewRedisClient,
		redispubsub.NewPublisher,
		redispubsub.NewSubscriber,
		sse.NewManager,
		service.NewOrdersService,
		wire.Bind(new(handler.OrdersServicer), new(*service.OrdersService)),
		router.NewRouter,
		NewApplication,
	)
	return &Application{}, func() {}, nil
}

func NewConfigApp() *config.App {
	return config.NewApp("config/config.yaml")
}

func NewPostgresConfig(appConfig *config.Config) *postgres.Config {
	return &postgres.Config{
		Host: appConfig.Db.Host,
		Port: appConfig.Db.Port,
		User: appConfig.Db.User,
		Pass: appConfig.Db.Pass,
		Name: appConfig.Db.Name,
	}
}

func NewRedisConfig(appConfig *config.Config) *redispubsub.Config {
	return &redispubsub.Config{
		Host:    appConfig.Redis.Host,
		Port:    appConfig.Redis.Port,
		Channel: appConfig.Redis.Channel,
	}
}

func NewRedisClient(redisConfig *redispubsub.Config) (*redis.Client, func(), error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		client.Close()
	}

	return client, cleanup, nil
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
	rtr *router.Router,
	cfg *config.Config,
	outboxPub *kafka.OutboxPublisher,
	inboxProc *kafka.InboxProcessor,
	ordSvc *service.OrdersService,
	sseMgr *sse.Manager,
) *Application {
	return &Application{
		Router:          rtr,
		Config:          cfg,
		OutboxPublisher: outboxPub,
		InboxProcessor:  inboxProc,
		OrdersService:   ordSvc,
		SSEManager:      sseMgr,
	}
}
