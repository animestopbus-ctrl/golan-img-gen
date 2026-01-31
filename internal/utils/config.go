package utils

import (
	"log"
	"os"
)

type Config struct {
	BotToken      string
	MongoURI      string
	PythonAPIURL  string
}

func LoadConfig() Config {
	return Config{
		BotToken:     getEnv("BOT_TOKEN", ""),
		MongoURI:     getEnv("MONGO_URI", ""),
		PythonAPIURL: getEnv("PYTHON_API_URL", ""),
	}
}

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		if defaultVal == "" {
			log.Fatalf("%s env var required", key)
		}
		return defaultVal
	}
	return val

}
