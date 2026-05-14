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
	"golang.org/x/oauth2"
)

type ChatRequest struct {
	Question      string `json:"question"`
	SourceContext string `json:"sourceContext"`
	GoogleAccessToken string `json:"googleAccessToken"`

}

type ChatResponse struct {
	Reply string `json:"reply"`
}

type accessTokenAuth struct {
	accessToken string
}

func (a accessTokenAuth) HTTPClient(ctx context.Context) (*http.Client, error) {
	if a.accessToken == "" {
		return nil, fmt.Errorf("google access token is required")
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: a.accessToken,
		TokenType:   "Bearer",
	})

	return oauth2.NewClient(ctx, tokenSource), nil
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

func logModelResponse(resp *responses.Response) {
	log.Printf("openai response id=%s output_text=%q output_items=%d", resp.ID, resp.OutputText(), len(resp.Output))
	for _, item := range resp.Output {
		if item.Type == "function_call" {
			log.Printf("openai tool_call name=%s call_id=%s arguments=%s", item.Name, item.CallID, item.Arguments.OfString)
			continue
		}

		log.Printf("openai output item type=%s", item.Type)
	}
}

func callModelWithToolOutput(
	ctx context.Context,
	client openai.Client,
	previousResponseID string,
	callID string,
	toolResultJSON string,
	modelTools []responses.ToolUnionParam,
) (*responses.Response, error) {
	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model:              openai.ChatModelGPT5_2,
		PreviousResponseID: openai.String(previousResponseID),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfFunctionCallOutput(callID, toolResultJSON),
			},
		},
		Tools: modelTools,
	})
	if err != nil {
		return nil, err
	}

	logModelResponse(resp)
	return resp, nil
}

func ask(ctx context.Context, question string, auth tool.HTTPClientProvider) (string, error) {
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

	for {
		functionCallHandled := false

		for _, item := range resp.Output {
			if item.Type != "function_call" {
				continue
			}

			functionCallHandled = true
			toolName := item.Name
			rawArgs := []byte(item.Arguments.OfString)

			toolResult, err := registery.Call(toolName, ctx, rawArgs, auth)
			if err != nil {
				return "", err
			}

			toolResultJSON, err := json.Marshal(toolResult)
			if err != nil {
				return "", err
			}

			log.Printf("tool_result name=%s output=%s", toolName, string(toolResultJSON))

			resp, err = callModelWithToolOutput(
				ctx,
				openai_client,
				resp.ID,
				item.CallID,
				string(toolResultJSON),
				modelTools,
			)
			if err != nil {
				return "", err
			}

			break
		}

		if !functionCallHandled {
			//	llm is finished calling tools
			if resp.OutputText() == "" {
				return "", fmt.Errorf("model returned no output text and no function call")
			}
			return resp.OutputText(), nil

		}
	}
}

func NewChatHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		auth := accessTokenAuth{
			accessToken: req.GoogleAccessToken,
		}

		gpt_resp, err := ask(r.Context(), req.Question, auth)

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
}
