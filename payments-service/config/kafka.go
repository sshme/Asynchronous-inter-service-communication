package config

import (
	"os"
	"strings"
)

type KafkaConfig struct {
	Brokers []string
	Topics  KafkaTopics
}

type KafkaTopics struct {
	OrdersEvents   string
	PaymentsEvents string
}

func LoadKafkaConfig() *KafkaConfig {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}

	brokers := strings.Split(brokersEnv, ",")
	for i, broker := range brokers {
		brokers[i] = strings.TrimSpace(broker)
	}

	return &KafkaConfig{
		Brokers: brokers,
		Topics: KafkaTopics{
			OrdersEvents:   "orders-events",
			PaymentsEvents: "payments-events",
		},
	}
}
