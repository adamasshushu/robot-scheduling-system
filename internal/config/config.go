package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB      DBConfig
	Redis   RedisConfig
	MQTT    MQTTConfig
	JWT     JWTConfig
	Server  ServerConfig
	MinIO   MinIOConfig
	DevMode bool
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type MQTTConfig struct {
	Broker   string
	Port     string
	Username string
	Password string
}

func (c MQTTConfig) URI() string {
	return fmt.Sprintf("tcp://%s:%s", c.Broker, c.Port)
}

type JWTConfig struct {
	Secret        string
	Expiry        string
	RefreshExpiry string
}

type ServerConfig struct {
	APIPort       string
	WebSocketPort string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
}

func Load() *Config {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "rss"),
			Password: getEnv("DB_PASSWORD", "rss_dev_2026"),
			Name:     getEnv("DB_NAME", "robot_scheduling"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		MQTT: MQTTConfig{
			Broker:   getEnv("MQTT_BROKER", "localhost"),
			Port:     getEnv("MQTT_PORT", "1883"),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "rss-dev-secret"),
			Expiry:        getEnv("JWT_EXPIRY", "2h"),
			RefreshExpiry: getEnv("REFRESH_TOKEN_EXPIRY", "168h"),
		},
		Server: ServerConfig{
			APIPort:       getEnv("API_PORT", "8000"),
			WebSocketPort: getEnv("WEBSOCKET_PORT", "8001"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minio_dev_2026"),
			Bucket:    getEnv("MINIO_BUCKET", "robot-assets"),
		},
		DevMode: getEnv("DEV_MODE", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
