# Конструирование программного обеспечения
[![Go CI](https://github.com/sshme/Asynchronous-inter-service-communication/actions/workflows/ci.yml/badge.svg)](https://github.com/sshme/Asynchronous-inter-service-communication/actions/workflows/ci.yml)

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
    subgraph "User & Client"
        User
        Client["React SPA"]
    end

    subgraph "Gateway & Infrastructure"
        ApiGateway["API Gateway<br/>Traefik"]
        Kafka["Apache Kafka<br/>Topics: orders-events, payments-events"]
        Redis["Redis<br/>Pub/Sub"]
    end

    subgraph "Services"
        OrdersService["Orders Service<br/>(Multiple Replicas)"]
        PaymentsService["Payments Service"]
    end

    subgraph "Databases & Tables"
        OrdersDB["Orders DB<br/>(orders, outbox, inbox)"]
        PaymentsDB["Payments DB<br/>(payments, outbox, inbox)"]
    end
    
    User --> Client
    Client -->|REST API| ApiGateway
    Client -.->|SSE Connection| OrdersService
    
    ApiGateway -->|/orders-api/| OrdersService
    ApiGateway -->|/payments-api/| PaymentsService

    OrdersService -- "Write Order & Outbox msg (atomic)" --> OrdersDB
    OrdersService -- "Publish 'order.created'" --> Kafka
    
    Kafka -- "Consume 'order.created'" --> PaymentsService
    PaymentsService -- "Process payment<br/>Write Payment & Outbox msg (atomic)" --> PaymentsDB
    PaymentsService -- "Publish 'payment.completed'" --> Kafka

    Kafka -- "Consume 'payment.completed'<br/>(any replica)" --> OrdersService
    OrdersService -- "Update Order in DB" --> OrdersDB
    OrdersService -- "Publish status to Redis" --> Redis
    
    Redis -- "SUBSCRIBED by ALL replicas" --> OrdersService
    OrdersService -- "Send update via SSE<br/>(only replica with connection)" --> Client
    
    classDef userClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px,color:#000
    classDef serviceClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    classDef dbClass fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px,color:#000
    classDef infraClass fill:#fff3e0,stroke:#e65100,stroke-width:2px,color:#000
    
    class User,Client userClass
    class OrdersService,PaymentsService serviceClass
    class OrdersDB,PaymentsDB dbClass
    class ApiGateway,Kafka,Redis infraClass
```

### Быстрые команды

**Orders service <br> Payments service**
```sh
wire ./... # di gen
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out # test
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
