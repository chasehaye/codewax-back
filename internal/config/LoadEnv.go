package config

import (
	"log"
	"os"
)

type Config struct {
    Env           string
    Port          string
    FrontendURL   string
    FrontendURL2  string
    DBHost        string
    DBPort        string
    DBUser        string
    DBPass        string
    AdminEmail    string
    JWTSecret     string
    SenderAddress string
    AWSAccessKey  string
    AWSSecretKey  string
    AWSRegion     string
    Voyage_API_Key string
}

var envIsProduction bool

func IsProduction() bool {
    return envIsProduction
}

func CheckRequiredEnvVarsAndLoad() *Config {
	required := []string{
		"GO_ENV",
		"POSTGRESQL_HOST",
		"POSTGRESQL_PASS",
		"POSTGRESQL_USER",
		"POSTGRESQL_PORT",
		"FRONTEND_URL",
		"ADMIN_EMAIL",
		"JWT_SECRET",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_DEFAULT_REGION",
		"SENDER_ADDRESS",
		"PORT",
        "VOYAGE_API_KEY",
        
	}
	
	for _, v := range required {
		if os.Getenv(v) == "" {
			log.Fatalf("CRITICAL CONFIG ERROR: Environment variable '%s' is missing. Server cannot start.", v)
		}
	}
	
	log.Println("Environment validation successful: All variables present.")
	
    env := os.Getenv("GO_ENV")
    switch env {
    case "prod", "production":
        envIsProduction = true
    default:
        envIsProduction = false
    }

	return &Config{
        Env:           env,
        Port:          os.Getenv("PORT"),
        FrontendURL:   os.Getenv("FRONTEND_URL"),
        DBHost:        os.Getenv("POSTGRESQL_HOST"),
        DBPort:        os.Getenv("POSTGRESQL_PORT"),
        DBUser:        os.Getenv("POSTGRESQL_USER"),
        DBPass:        os.Getenv("POSTGRESQL_PASS"),
        AdminEmail:    os.Getenv("ADMIN_EMAIL"),
        JWTSecret:     os.Getenv("JWT_SECRET"),
        SenderAddress: os.Getenv("SENDER_ADDRESS"),
        AWSAccessKey:  os.Getenv("AWS_ACCESS_KEY_ID"),
        AWSSecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
        AWSRegion:     os.Getenv("AWS_DEFAULT_REGION"),
        Voyage_API_Key: os.Getenv("VOYAGE_API_KEY"),
    }
}