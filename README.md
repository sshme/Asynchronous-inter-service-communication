# Конструирование программного обеспечения
> Контрольная работа №3 <br> Асинхронное межсервисное взаимодействие.

## Описание системы

Система реализует микросервисную архитектуру с асинхронным межсервисным взаимодействием через Apache Kafka, используя паттерны **Inbox** и **Outbox** для обеспечения надежности доставки сообщений и eventual consistency, **exactly once** для оплаты заказа.

## Особенности реализации

1. Использование **SSE** (server-sent events) вместо WebSockets:
    1. Одностороннее общение
    2. Проще в использовании (Автоматическое переподключение)

2. Использование uuid v7 для id сущностей (часть доменного слоя).

3. Domain driven design

## Функционал

1. При инициализации клиентского приложения осуществляется запрос на создание пользователя (user id сохраняется в localStorage), также можно выйти из аккаунт и создать нового пользователя (кнопка logout).
2. Кнопка создания заказа (стоимость генерируется в orders-service).
3. Кнопка пополнения аккаунта (на 100 у.е.).
4. Клиент подписывается на изменения заказов и отслеживает изменения статусов заказов в реальном времени.

## Схема работы
```mermaid
graph TB
    User
    Client["React SPA"]
    Dashboard["Dashboard"]
    
    ApiGateway["API Gateway<br/>Traefik"]
    
    OrdersService["Orders Service"]
    PaymentsService["Payments Service"]

    OrdersSwagger["swagger docs"]
    PaymentsSwagger["swagger docs"]
    
    OrdersDB["Orders DB"]
    PaymentsDB["Payments DB"]
    
    OrdersOutbox["Orders Outbox"]
    PaymentsOutbox["Payments Outbox"]
    
    OrdersInbox["Orders Inbox<br/>Exactly-Once"]
    PaymentsInbox["Payments Inbox<br/>Exactly-Once"]
    
    DeduplicationLayer["Deduplication Layer<br/>EventID uniqueness"]
    IdempotencyLayer["Idempotency Layer<br/>Check existing payments"]
    
    Kafka["Apache Kafka<br/>Topics: orders-events, payments-events<br/>EventID per message"]
    Zookeeper["Zookeeper"]
    KafkaUI["Kafka UI"]
    
    User --> Client
    User --> ApiGateway
    Client --> ApiGateway
    
    ApiGateway -->|"/orders-api/*"| OrdersService
    ApiGateway -->|"/payments-api/*"| PaymentsService
    ApiGateway -->|"/dashboard"| Dashboard
    
    OrdersService --> OrdersDB
    PaymentsService --> PaymentsDB

    PaymentsService -->|"/payments-api/docs"| PaymentsSwagger
    OrdersService -->|"/orders-api/docs"| OrdersSwagger
    
    OrdersOutbox --> OrdersDB
    OrdersOutbox -->|"Publish Events<br/>with unique EventID<br/>(order.created,<br/>order.updated,<br/>order.completed)"| Kafka
    
    PaymentsOutbox --> PaymentsDB  
    PaymentsOutbox -->|"Publish Events<br/>with unique EventID<br/>(payment.completed,<br/>payment.failed)"| Kafka
    
    Kafka -->|"IsEventProcessed()"| PaymentsInbox
    PaymentsInbox -->|"UNIQUE constraint"| DeduplicationLayer
    DeduplicationLayer -->|"Event not processed"| PaymentsDB
    PaymentsInbox -->|"Idempotent Check<br/>GetByOrderID()"| IdempotencyLayer
    IdempotencyLayer -->|"Check existing payment"| PaymentsDB
    IdempotencyLayer -->|"Create or use existing"| PaymentsService
    
    Kafka -->|"IsEventProcessed()"| OrdersInbox
    OrdersInbox -->|"UNIQUE constraint"| DeduplicationLayer
    DeduplicationLayer -->|"Event not processed"| OrdersDB
    OrdersInbox -->|"Idempotent Processing"| OrdersService
    OrdersService -->|"Atomic Transaction<br/>Order + Outbox + Status"| OrdersDB
    
    Kafka -.->|"Cluster<br/>Coordination"| Zookeeper
    KafkaUI -.-> Kafka
    
    DeduplicationLayer -.->|"Prevents duplicate<br/>processing"| IdempotencyLayer
    IdempotencyLayer -.->|"Safe re-execution"| PaymentsService
    
    classDef userClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px,color:#000
    classDef serviceClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    classDef dbClass fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef kafkaClass fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    classDef outboxClass fill:#fce4ec,stroke:#880e4f,stroke-width:2px,color:#000
    classDef inboxClass fill:#e3f2fd,stroke:#0d47a1,stroke-width:2px,color:#000
    classDef exactlyOnceClass fill:#fff8e1,stroke:#f57c00,stroke-width:3px,color:#000
    
    class User,Client,Dashboard userClass
    class ApiGateway,OrdersService,PaymentsService serviceClass
    class OrdersDB,PaymentsDB dbClass
    class Kafka,Zookeeper,KafkaUI kafkaClass
    class OrdersOutbox,PaymentsOutbox outboxClass
    class OrdersInbox,PaymentsInbox inboxClass
    class DeduplicationLayer,IdempotencyLayer exactlyOnceClass
```

### Быстрые команды

**Orders service <br> Payments service**
```sh
wire ./... # di gen
test ./... # test
swag init -g cmd/api/main.go # docs gen
```

**Orders client**
```sh
pnpm i # deps install
pnpm run dev # run in dev mode
```

**Запуск**
```sh
docker-compose up
```

**Запуск в minikube**
> Запустит 3 реплики <i>orders-service</i> и <i>payments-service</i>
```sh
minikube start
./k8s/build-images.sh
./k8s/deploy.sh
minikube tunnel
```
