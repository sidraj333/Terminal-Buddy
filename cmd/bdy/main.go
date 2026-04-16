package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"terminal-buddy/internal/backend"
	gauth "terminal-buddy/internal/backend/google"
)





func main() {
	fmt.Println("Buddy new session started.")
	fmt.Println("Type /bye or /exit to leave.")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Unable to resolve home directory:", err)
		return
	}

	credsPath := "credentials.json"
	tokenPath := filepath.Join(homeDir, ".config", "terminal-buddy", "token.json")
	scopes := []string{
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/documents",
		"https://www.googleapis.com/auth/spreadsheets",
		"https://www.googleapis.com/auth/presentations",
	}

	auth := gauth.NewAuthManager(credsPath, tokenPath, scopes)
	if err := auth.Login(context.Background()); err != nil {
		fmt.Println("Google login failed:", err)
		return
	}

	

	verbose := flag.Bool("verbose", false, "show debug logs")
	flag.Parse()

	var logger *log.Logger
	if *verbose {
		logger = log.New(os.Stderr, "[LOG] ", log.LstdFlags)
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("bdy> ")

		if !scanner.Scan() {
			fmt.Println("\nBuddy session ended.")
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		cmd := strings.ToLower(input)
		if cmd == "/bye" || cmd == "/exit" {
			fmt.Println("goodbye :)")
			return
		}

		if strings.HasPrefix(input, "/doc_read") {
			docURL := strings.TrimSpace(strings.TrimPrefix(input, "/doc_read "))
			document := gauth.NewDocService(context.Background(), docURL, auth)
			_, err := document.GetDoc()
			if err != nil {
				fmt.Println(err)
				fmt.Println("coudl not fetch doc")
			}
			continue
		}

		fmt.Println("thinking...")
		reply, err := backend.NewAIService(logger).Reply(context.Background(), input)
		if err != nil {
			logger.Println(err)
			fmt.Println("An error occurred while processing your request. Please try again.")
			continue
		}


		fmt.Print(reply)
	}
}
