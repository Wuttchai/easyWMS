package config

import "os"

type Config struct{ AppPort, DBHost, DBPort, DBName, DBUser, DBPassword, DBSSLMode, JWTSecret string }

func Load() Config {
	return Config{AppPort: getEnv("APP_PORT", "8080"), DBHost: getEnv("DB_HOST", "localhost"), DBPort: getEnv("DB_PORT", "5432"), DBName: getEnv("DB_NAME", "easywms"), DBUser: getEnv("DB_USER", "easywms"), DBPassword: getEnv("DB_PASSWORD", "easywms123"), DBSSLMode: getEnv("DB_SSLMODE", "disable"), JWTSecret: getEnv("JWT_SECRET", "easywms-demo-secret-change-me")}
}
func getEnv(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
