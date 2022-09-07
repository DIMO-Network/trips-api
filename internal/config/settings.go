package config

type Settings struct {
	DBUser       string `yaml:"DB_USER"`
	DBPassword   string `yaml:"DB_PASSWORD"`
	DBPort       string `yaml:"DB_PORT"`
	DBHost       string `yaml:"DB_HOST"`
	DBName       string `yaml:"DB_NAME"`
	KafkaBrokers string `yaml:"KAFKA_BROKERS"`
}
