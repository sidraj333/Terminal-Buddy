 package main

  import (
  	"bufio"
  	"fmt"
  	"os"
  	"strings"
  )

  func main() {
  	fmt.Println("Buddy session started.")
  	fmt.Println("Type /bye or /exit to leave.")

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
  			fmt.Println("Buddy session ended.")
  			return
  		}

  		// Placeholder app behavior.
  		fmt.Printf("You said: %s\n", input)
  	}
  }