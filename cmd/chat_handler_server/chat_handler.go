package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	tool "terminal-buddy/internal/tools"
	"terminal-buddy/internal/tools/google"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type ChatRequest struct {
	Question      string `json:"input"`
	SourceContext string
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func convertInputSchema(schema tool.InputSchema) map[string]any {
	properties := make(map[string]any)
	required := []string{}
	for _, param := range schema.Parameters {
		properties[param.Name] = map[string]any{
			"type": param.Type,
			"description": param.Description,
		}
		if param.Required {
			required = append(required, param.Name)
		}
	}
	return map[string]any{
		"type": "object",
		"properties": properties,
		"required": required,
		"additionalProperties": false,

	}
}

func buildModelTools(registery *tool.Registery) []responses.ToolUnionParam {
	modelTools := []responses.ToolUnionParam{}

	for _, registeredTool := range registery.ListTools() {
		modelTool := responses.ToolParamOfFunction(
			registeredTool.Name,
			convertInputSchema(registeredTool.InputSchema),
			true,
		)
		modelTool.OfFunction.Description = openai.String(registeredTool.Description)
		modelTools = append(modelTools, modelTool)
	}

	return modelTools
}

func callModelWithTools(
	ctx context.Context,
	client openai.Client,
	question string,
	modelTools []responses.ToolUnionParam,
) (*responses.Response, error) {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(question)},
		Model: openai.ChatModelGPT5_2,
		Tools: modelTools,
	})

	if err != nil {
		return nil, err
	}
	log.Printf("openai response id=%s output_text=%q output_items=%d", resp.ID, resp.OutputText(), len(resp.Output))
	for _, item := range resp.Output {
		if item.Type == "function_call" {
			log.Printf("openai tool_call name=%s call_id=%s arguments=%s", item.Name, item.CallID, item.Arguments.OfString)
			continue
		}

		log.Printf("openai output item type=%s", item.Type)
	}

	return resp, nil
}


func ask(ctx context.Context, question string) (string, error) {
	registery := tool.NewRegistery()
	if err := registery.RegisterTool("read_google_doc", google.NewReadGoogleDocTool()); err != nil {
		return "", err
	}

	modelTools := buildModelTools(registery)

	openai_client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	resp, err := callModelWithTools(ctx, openai_client, question, modelTools)
	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	gpt_resp, err := ask(r.Context(), req.Question)

	if err != nil {
		http.Error(w, fmt.Sprintf("ask failed: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ChatResponse{Reply: gpt_resp}
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
