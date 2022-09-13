package config

import "fmt"

// Settings credentials
type Settings struct {
	DBUser                         string `yaml:"DB_USER"`
	DBPassword                     string `yaml:"DB_PASSWORD"`
	DBPort                         string `yaml:"DB_PORT"`
	DBHost                         string `yaml:"DB_HOST"`
	DBName                         string `yaml:"DB_NAME"`
	KafkaBrokers                   string `yaml:"KAFKA_BROKERS"`
	JWTKeySetURL                   string `yaml:"JWT_KEY_SET_URL"`
	DevicesAPIGRPCAddr             string `yaml:"DEVICES_API_GRPC_ADDR"`
	ElasticSearchAnalyticsHost     string `yaml:"ELASTIC_SEARCH_ANALYTICS_HOST"`
	ElasticSearchAnalyticsUsername string `yaml:"ELASTIC_SEARCH_ANALYTICS_USERNAME"`
	ElasticSearchAnalyticsPassword string `yaml:"ELASTIC_SEARCH_ANALYTICS_PASSWORD"`
	ElasticSearchIndex             string `yaml:"ELASTIC_INDEX"`
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
