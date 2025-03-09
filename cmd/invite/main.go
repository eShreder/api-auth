package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/user/user-server/pkg/auth"
	"github.com/user/user-server/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values or command line flags")
	}

	dbPath := flag.String("db", getEnv("DB_PATH", "user-server.db"), "Path to SQLite database file")
	flag.Parse()

	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	if err := db.Initialize(); err != nil {
		log.Fatalf("Database initialization error: %v", err)
	}

	token, err := auth.GenerateInviteToken()
	if err != nil {
		log.Fatalf("Invite token generation error: %v", err)
	}

	inviteToken, err := db.CreateInviteToken(token)
	if err != nil {
		log.Fatalf("Invite token save error: %v", err)
	}

	fmt.Printf("New invite token created: %s\n", inviteToken.Token)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
} 