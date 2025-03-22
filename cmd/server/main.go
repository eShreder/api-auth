package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/user/user-server/pkg/auth"
	"github.com/user/user-server/pkg/database"
	"github.com/user/user-server/pkg/handlers"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values or command line flags")
	}

	dbPath := flag.String("db", getEnv("DB_PATH", "user-server.db"), "Path to SQLite database file")
	privateKeyPath := flag.String("private-key", getEnv("PRIVATE_KEY_PATH", "keys/private.pem"), "Path to RSA private key")
	publicKeyPath := flag.String("public-key", getEnv("PUBLIC_KEY_PATH", "keys/public.pem"), "Path to RSA public key")
	addr := flag.String("addr", getEnv("ADDR", ":8080"), "HTTP listen address")
	jwtTTLHours := flag.Int("jwt-ttl", getEnvAsInt("JWT_TTL", 24), "JWT token lifetime in hours")
	flag.Parse()

	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}

	jwtManager, err := auth.NewJWTManager(*privateKeyPath, *publicKeyPath, time.Duration(*jwtTTLHours)*time.Hour)
	if err != nil {
		log.Fatalf("JWT manager initialization error: %v", err)
	}

	authHandler := &handlers.AuthHandler{
		DB:         db,
		JWTManager: jwtManager,
	}

	router := gin.Default()

	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/refresh", authHandler.Refresh)
	router.GET("/api/auth/public-key", authHandler.GetPublicKey)

	protected := router.Group("/api")
	protected.Use(authHandler.AuthMiddleware())
	{
		protected.GET("/me", authHandler.GetMe)
	}

	log.Printf("Server started on %s", *addr)
	if err := router.Run(*addr); err != nil {
		log.Fatalf("Server start error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
} 