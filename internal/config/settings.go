package config

import (
	"github.com/DIMO-Network/shared/db"
)

// Settings credentials
type Settings struct {
	Port    string `yaml:"PORT"`
	MonPort string `yaml:"MON_PORT"`

	DB db.Settings `yaml:"DB"`
	// DBUser     string      `yaml:"DB_USER"`
	// DBPassword string      `yaml:"DB_PASSWORD"`
	// DBPort     string      `yaml:"DB_PORT"`
	// DBHost     string      `yaml:"DB_HOST"`
	// DBName     string      `yaml:"DB_NAME"`

	ElasticHost     string `yaml:"ELASTIC_HOST"`
	ElasticUsername string `yaml:"ELASTIC_USERNAME"`
	ElasticPassword string `yaml:"ELASTIC_PASSWORD"`
	ElasticIndex    string `yaml:"ELASTIC_INDEX"`

	KafkaBrokers     string `yaml:"KAFKA_BROKERS"`
	TripEventTopic   string `yaml:"TRIP_EVENT_TOPIC"`
	BundlrPrivateKey string `yaml:"BUNDLR_PRIVATE_KEY"`
	BundlrNetwork    string `yaml:"BUNDLR_NETWORK"`
	BundlrCurrency   string `yaml:"BUNDLR_CURRENCY"`
	EventTopic       string `yaml:"EVENTS_TOPIC"`

	DataFetchEnabled bool `yaml:"DATA_FETCH_ENABLED"`
	WorkerCount      int  `yaml:"WORKER_COUNT"`
	BundlrEnabled    bool `yaml:"BUNDLR_ENABLED"`

	PrivilegeJWKURL string `yaml:"PRIVILEGE_JWK_URL"`

	VehicleNFTAddr string `yaml:"VEHICLE_NFT_ADDR"`
}
