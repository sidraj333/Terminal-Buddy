package backend

import (
	"context"
	"log"

)

type AIService struct {
	logger *log.Logger
}

func NewAIService(logger *log.Logger) *AIService {
	return &AIService{
		logger: logger,
	}
}


func (s *AIService) Reply(ctx context.Context, userInput string) (string, error) {
	s.logger.Printf("AIService received input: %s\n", userInput);
	
	return "Placeholder", nil;
}


