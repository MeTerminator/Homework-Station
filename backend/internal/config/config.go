package config

import (
	"os"
	"strings"
)

type Database struct {
	Driver   string
	DSN      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Config struct {
	Port           string
	DataDir        string
	SessionSecret  string
	BindingSecret  string
	CookieSecure   bool
	TrustedProxies []string
	Database       Database
}

func Or(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

// Load snapshots the process environment once; downstream layers never read it.
func Load() Config {
	trustedProxies, configured := os.LookupEnv("TRUSTED_PROXIES")
	if !configured {
		trustedProxies = "127.0.0.1"
	}
	return Config{
		Port:          Or(os.Getenv("PORT"), "8080"),
		DataDir:       Or(os.Getenv("DATA_DIR"), "./data"),
		SessionSecret: os.Getenv("JWT_SECRET"),
		BindingSecret: os.Getenv("BINDING_SECRET"),
		CookieSecure:  os.Getenv("COOKIE_SECURE") == "true",
		TrustedProxies: strings.FieldsFunc(trustedProxies, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}),
		Database: Database{
			Driver:   Or(os.Getenv("DB_DRIVER"), "sqlite"),
			DSN:      os.Getenv("DB_DSN"),
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     Or(os.Getenv("DB_NAME"), "homework"),
			SSLMode:  Or(os.Getenv("DB_SSLMODE"), "disable"),
		},
	}
}
