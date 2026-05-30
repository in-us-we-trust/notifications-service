package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Environment constants for application mode.
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
)

// LoadEnv loads environment variables from the .env file.
// In production (Cloud Run), the .env file doesn't exist, so we ignore the error.
func LoadEnv() error {
	// prueba la raíz primero, luego el path relativo para cuando se corre desde cmd/buen
	for _, path := range []string{".env", "../../.env"} {
		if err := godotenv.Load(path); err == nil {
			return nil
		}
	}
	return nil
}

func GetString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return valAsInt
}

func GetBool(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valAsBool, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}

	return valAsBool
}
