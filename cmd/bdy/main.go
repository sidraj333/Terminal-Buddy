 package main

  import (
  	"bufio"
  	"fmt"
  	"os"
  	"strings"
	"terminal-buddy/internal/backend"
	"context"
	"log"
	"io"
	"flag"
  )

  func main() {
  	fmt.Println("Buddy session started.")
  	fmt.Println("Type /bye or /exit to leave.")

	verbose := flag.Bool("verbose", false, "show debug logs")
	// flag returns pointers to booleans
	
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

  		// If user sends EOF (Ctrl+D), end session and return to shell.
  		if !scanner.Scan() {
  			fmt.Println("\nBuddy session ended.")
  			return
  		}

  		input := strings.TrimSpace(scanner.Text())
  		if input == "" {
  			continue
  		}

  		cmd := strings.ToLower(input)

  		// Exit commands end this process.
  		if cmd == "/bye" || cmd == "/exit" {
  			fmt.Println("goodbye :)")
  			return
  		}
		fmt.Println("thinking...")
		// call the function here
		reply, error := backend.NewAIService(logger).Reply(context.Background(), input)

		if error != nil{
			logger.Println(error)
			fmt.Println("An error occurred while processing your request. Please try again.")
			continue
		}


  		// Placeholder app behavior.
  		fmt.Print(reply)
  	}
  }