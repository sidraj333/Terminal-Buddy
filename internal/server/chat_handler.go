package server

import (
	"encoding/json"
	"net/http"
	"context"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"os"
)

type ChatRequest struct{
	Input string `json:"input"`
} 

type ChatResponse struct{
	Reply string `json:"reply"` 
}



func ask(ctx context.Context, question string) (string, error) {
	//helper function to call gpt api for a question

	openai_client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	resp, err := openai_client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(question)},
		Model: openai.ChatModelGPT5_2,
	})

	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	gpt_resp, err := ask(r.Context(), req.Input)

	if err != nil {
		http.Error(w, "400", http.StatusInternalServerError)
	}

	resp := ChatResponse{Reply: gpt_resp}
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
  	}
}