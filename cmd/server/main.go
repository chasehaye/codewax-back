// @title           Log Relay API
// @version         1.0
// @description     API Server for Log Relay Project.
// @host            localhost:8080
// @BasePath        /
package main

import (
    "log"
    "os"


    "github.com/joho/godotenv"

    "codewax/internal/database"
    "codewax/internal/models"
    "codewax/internal/config"
    "codewax/internal/router"
)

func main() {
    // ------------------ ENV VALIDATION ------------------
    _ = godotenv.Load() 
    cfg := config.CheckRequiredEnvVarsAndLoad()

    // ------------------ DATABASE ------------------
    db, err := database.ConnectToDB(cfg.DBHost, cfg.DBUser, cfg.DBPass, "code_wax", cfg.DBPort)
    if err != nil {
        log.Fatalln("Failed to connect to database:", err)
    }
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalln("Failed to get generic database object:", err)
    }
    if err := sqlDB.Ping(); err != nil {
        log.Fatalln("Database unreachable:", err)
    }
    defer sqlDB.Close()
    log.Println("--Database connection verified--------------------------------------")

    // ------------------ Migrate ------------------
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
        log.Printf("failed to create vector extension: %v", err)
    }

    err = db.AutoMigrate(
        &models.User{},
        &models.PasswordReset{},
        &models.EmailChangeRequest{},
        &models.Conversation{},
        &models.Message{},
        &models.Repository{},
        &models.RepositoryChunk{},
    )
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    log.Println("--Database migration successful-------------------------------------")
    log.Println("--Connected---------------------------------------------------------")

    // ------------------ START SERVER ------------------
    r := router.Setup(db)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on port %s in %s mode...", cfg.Port, cfg.Env)
    r.Run(":" + port)

}