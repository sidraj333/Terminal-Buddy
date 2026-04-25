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
	gauth "terminal-buddy/internal/backend/google"

)

type Source interface{
	Write() error 
	Read() error
	Type() error
	Ask(question string) (string, error)
}

var source Source = nil // this is the global object that handles calls to google drive api


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
	//prompt struct (doc struct, presentation struct, excel struct)
	

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
		logger.Println("starting doc fetch")
		
		// set a string here we write type
		// 3 variables here
		if strings.HasPrefix(input, "/open") {	
			// TODO  parse the url and implement checks to determine what type of source
			
			URL := strings.TrimSpace(strings.TrimPrefix(input, "/open "))
			source, err := gauth.NewDocService(context.Background(), URL, auth)
			if err != nil {
				fmt.Println(err)
				fmt.Println("Could not fetch doc")
				continue
			}
			source.Read()

			fmt.Println("Successful doc fetch, waiting for your questions")

		} else {
			fmt.Println("thinking...")
			
			gtp_resp, err := source.Ask(input)
			if err != nil {
				fmt.Println("ERROR calling gpt")
			} else {
				fmt.Println(gtp_resp)
			}
		
		}



		


		
	}
}



