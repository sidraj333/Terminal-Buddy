package backend

import (
	"context"
	"log"
	"strings"
	"time"
	"os"
	"net/http"
	"fmt"
	"encoding/json"
	"bytes"
	"io"
	"errors"
	
)

type AIService struct {
	logger *log.Logger
	httpClient *http.Client
	apiKey string
	model string
}

func NewAIService(logger *log.Logger) *AIService {
	return &AIService{
		logger: logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey: strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		model: "gpt-4.1-mini",
	}
}


func (s *AIService) Reply(ctx context.Context, userInput string) (string, error) {
	/*
		function responsible from taking user input and making http call to Open AI API
		and sending the AI response back to the user
	*/

	s.logger.Printf("AIService received input: %s\n", userInput);

	reqBody := map[string]any {
		"model": s.model,
		"input": userInput,
	}

	s.logger.Printf("Sending request to Open AI")

	reqBodyBytes, err := json.Marshal(reqBody);

	if err != nil {
		s.logger.Printf("Error marshaling request body: %v\n", err)
		return "", fmt.Errorf("marshal request: %w", err)
	}


	//create http object to send request to Open AI
	http_req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/responses",
		bytes.NewReader(reqBodyBytes),

	)

	if err != nil {
		s.logger.Printf("Error creating HTTP request: %v\n", err)
		return "", fmt.Errorf("create request: %w", err)
	}

	http_req.Header.Set("Authorization", "Bearer "+s.apiKey)
  	http_req.Header.Set("Content-Type", "application/json")

	http_resp, err := s.httpClient.Do(http_req)
	if err != nil {
		s.logger.Printf("Error making HTTP request: %v\n", err)
		return "", fmt.Errorf("make request: %w", err)
	}

	defer http_resp.Body.Close() // close http connection when function is done running

	http_body, err := io.ReadAll(http_resp.Body)
	s.logger.Printf("Raw OpenAI response: %s", string(http_body))

	if http_resp.StatusCode < 200 || http_resp.StatusCode >= 300 {
		s.logger.Printf("Received non-2xx response: %d\n", http_resp.StatusCode)
		s.logger.Printf("OpenAI error status=%d body=%s", http_resp.StatusCode, string(http_body))
		return "", fmt.Errorf("There was an error calling open ai")
	}

	var parsed_http_response struct {
		OutputText string `json:"output_text"`
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(http_body, &parsed_http_response); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	out := strings.TrimSpace(parsed_http_response.OutputText)

	if out == "" {
		var builder strings.Builder
		for _, item := range parsed_http_response.Output {
			for _, content := range item.Content {
				if content.Type == "output_text" && strings.TrimSpace(content.Text) != "" {
					if builder.Len() > 0 {
						builder.WriteString("\n")
					}
					builder.WriteString(content.Text)
				}
			}
		}
		out = strings.TrimSpace(builder.String())
	}
	
  	if out == "" {
  		return "", errors.New("empty model response")
  	}

  	s.logger.Printf("OpenAI reply received, len=%d", len(out))
  	return out, nil


}

