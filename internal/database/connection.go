package database

import (
    "fmt"
    "log"
    "os"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func ConnectToDB(dbHost, dbUser, dbPassword, dbName, dbPort string) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", 
        dbHost, dbUser, dbPassword, dbName, dbPort)

        env := os.Getenv("GO_ENV")

    var logLevel logger.LogLevel
    var ignoreNotFound bool
    var colorful bool

    if env == "prod" {
        logLevel = logger.Error
        ignoreNotFound = true
        colorful = false
    } else {
        logLevel = logger.Warn
        ignoreNotFound = false
        colorful = true
    }

    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logLevel,
            IgnoreRecordNotFoundError: ignoreNotFound,
            Colorful:                  colorful,
        },
    )

    return gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })
}
