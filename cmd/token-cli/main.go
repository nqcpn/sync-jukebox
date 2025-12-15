package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gofrs/uuid"
	"github.com/yeeeck/sync-jukebox/internal/db"
)

const dbPath = "./jukebox.db"

func main() {
	action := flag.String("action", "", "Action to perform: generate, disable, enable")
	token := flag.String("token", "", "Token to act upon for disable/enable actions")
	flag.Parse()

	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	switch *action {
	case "generate":
		newToken, _ := uuid.NewV4()
		tokenStr := newToken.String()
		if err := database.AddToken(tokenStr); err != nil {
			log.Fatalf("Failed to generate token: %v", err)
		}
		fmt.Println("New token generated successfully:")
		fmt.Println(tokenStr)
	case "disable":
		if *token == "" {
			log.Fatal("'-token' flag is required for 'disable' action")
		}
		if err := database.SetTokenState(*token, false); err != nil {
			log.Fatalf("Failed to disable token %s: %v", *token, err)
		}
		fmt.Printf("Token %s has been disabled.\n", *token)
	case "enable":
		if *token == "" {
			log.Fatal("'-token' flag is required for 'enable' action")
		}
		if err := database.SetTokenState(*token, true); err != nil {
			log.Fatalf("Failed to enable token %s: %v", *token, err)
		}
		fmt.Printf("Token %s has been enabled.\n", *token)
	default:
		fmt.Println("Invalid action. Use 'generate', 'disable', or 'enable'.")
		flag.Usage()
		os.Exit(1)
	}
}
