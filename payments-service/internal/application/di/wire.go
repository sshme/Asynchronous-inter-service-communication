//go:build wireinject
// +build wireinject

package di

import (
	"database/sql"

	"payments-service/internal/application/service"
	"payments-service/internal/infrastructure/brokers/kafka"
	"payments-service/internal/infrastructure/config"
	"payments-service/internal/infrastructure/persistence/postgres"
	"payments-service/internal/interfaces/api/handler"
	"payments-service/internal/interfaces/api/router"
	"payments-service/internal/interfaces/repository"
	"payments-service/pkg/random"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	postgres.NewAccountRepository,
	postgres.NewPaymentsRepository,
	postgres.NewInboxRepository,
	postgres.NewOutboxRepository,
)

var RandomSet = wire.NewSet(
	random.NewCryptoGenerator,
	wire.Bind(new(random.Generator), new(*random.CryptoGenerator)),
)

var ServiceSet = wire.NewSet(
	service.NewPaymentsService,
	service.NewAccountService,
)

var HandlerSet = wire.NewSet(
	handler.NewAccountsHandler,
)

var KafkaSet = wire.NewSet(
	kafka.NewConfig,
	NewOutboxPublisher,
	NewInboxProcessor,
)

func InitializeApplication() (*Application, error) {
	wire.Build(
		NewConfigApp,
		config.MustLoad,
		NewPostgresConfig,
		RepositorySet,
		RandomSet,
		ServiceSet,
		HandlerSet,
		KafkaSet,
		router.NewRouter,
		postgres.NewDb,
		wire.Bind(new(service.DBTX), new(*sql.DB)),
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
	PaymentsService *service.PaymentsService
	AccountService  *service.AccountService
	OutboxPublisher *kafka.OutboxPublisher
	InboxProcessor  *kafka.InboxProcessor
	DB              *sql.DB
}

func NewApplication(
	router *router.Router,
	config *config.Config,
	paymentsService *service.PaymentsService,
	accountService *service.AccountService,
	outboxPublisher *kafka.OutboxPublisher,
	inboxProcessor *kafka.InboxProcessor,
	db *sql.DB,
) *Application {
	return &Application{
		Router:          router,
		Config:          config,
		PaymentsService: paymentsService,
		AccountService:  accountService,
		OutboxPublisher: outboxPublisher,
		InboxProcessor:  inboxProcessor,
		DB:              db,
	}
}
