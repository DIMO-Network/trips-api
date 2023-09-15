package config

import "fmt"

// Settings credentials
type Settings struct {
	Port    string `yaml:"PORT"`
	MonPort string `yaml:"MON_PORT"`

	DBUser     string `yaml:"DB_USER"`
	DBPassword string `yaml:"DB_PASSWORD"`
	DBPort     string `yaml:"DB_PORT"`
	DBHost     string `yaml:"DB_HOST"`
	DBName     string `yaml:"DB_NAME"`

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

	JWTKeySetURL     string `yaml:"JWT_KEY_SET_URL"`
	DataFetchEnabled bool   `yaml:"DATA_FETCH_ENABLED"`
	WorkerCount      int    `yaml:"WORKER_COUNT"`
	BundlrEnabled    bool   `yaml:"BUNDLR_ENABLED"`

	RPCUrl                     string `yaml:"RPC_URL"`
	DeployedContractPrivateKey string `yaml:"DEPLOYED_CONTRACT_PRIVATE_KEY"`
	DeployedContractAddress    string `yaml:"DEPLOYED_CONTRACT_ADDRESS"`
	UsersAPIGRPCAddr           string `yaml:"USERS_API_GRPC_ADDR"`
	ChainID                    int64  `yaml:"CHAIN_ID"`
}

// GetWriterDSN builds the connection string to the db writer - for now same as reader
func (app *Settings) GetWriterDSN(withSearchPath bool) string {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		app.DBUser,
		app.DBPassword,
		app.DBName,
		app.DBHost,
		app.DBPort,
	)
	if withSearchPath {
		dsn = fmt.Sprintf("%s search_path=%s", dsn, app.DBName) // assumption is schema has same name as dbname
	}
	return dsn
}
