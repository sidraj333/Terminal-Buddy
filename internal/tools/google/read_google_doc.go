package google

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"net/url"
	"strings"

	"terminal-buddy/internal/tools"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

var readGoogleDocInputSchema = tools.InputSchema{
	Parameters: []tools.ParameterSchema{
		{
			Name:        "document_url",
			Type:        "string",
			Description: "The Google Docs URL to read",
			Required:    true,
		},
	},
}

type ReadGoogleDocArgs struct {
	DocumentURL string `json:"document_url"`
}

type ReadGoogleDocResult struct {
	DocumentURL string         `json:"document_url"`
	Title       string         `json:"title"`
	Document    *docs.Document `json:"document"`
}

func NewReadGoogleDocTool() tools.Tool {
	return tools.Tool{
		Name:        "read_google_doc",
		Description: "Read a Google Doc from a Google Doc URL",
		InputSchema: readGoogleDocInputSchema,
		Handler:     ReadGoogleDocHandler,
	}
}

func GetDocumentId(document_url string) (string, error) {
	raw := strings.TrimSpace(document_url)
	if raw == "" {
		return "", errors.New("document url is empty")
	}

	u, err := url.Parse(document_url)
	if err != nil {
		return "", fmt.Errorf("invalid document url: %w", err)
	}

	if !strings.Contains(u.Host, "docs.google.com") {
		return "", fmt.Errorf("not a valid Google doc link: %s", raw)
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 4 || parts[1] != "document" || parts[2] != "d" || parts[3] == "" {
		return "", fmt.Errorf("could not extract doc id from url: %s", raw)
	}
	return parts[3], nil
}

func ReadGoogleDocHandler(ctx context.Context, rawArgs []byte, auth tools.HTTPClientProvider) (any, error) {
	var read_args ReadGoogleDocArgs
	if err := json.Unmarshal(rawArgs, &read_args); err != nil {
		return nil, err
	}
	
	if read_args.DocumentURL == "" {
		return nil, fmt.Errorf("expected document url as a parameter")
	}

	document_url := read_args.DocumentURL
	doc_id, err := GetDocumentId(document_url)
	if err != nil {
		return nil, err
	}

	if auth == nil {
		return nil, errors.New("auth manager is required")
	}

	http_client, err := auth.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}

	googleDocsHandler, err := docs.NewService(ctx, option.WithHTTPClient(http_client))
	if err != nil {
		return nil, err
	}

	document, err := googleDocsHandler.Documents.Get(doc_id).Do()
	if err != nil {
		return nil, err
	}

	return ReadGoogleDocResult{
		DocumentURL: read_args.DocumentURL,
		Title: document.Title,
		Document: document,
	}, nil
}
