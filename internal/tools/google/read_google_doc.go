package google

import (
	"context"
	"encoding/json"
	"fmt"
	backendgoogle "terminal-buddy/internal/backend/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
	"strings"
	"errors"
	"net/url"
)

type ReadGoogleDocsArgs struct {
	DocumentURL string `json:"document_url"`
}

type ReadGoogleDocResult struct {
	DocumentURL string `json:"document_url"`
	Title 		string `json:"title"`
	Document	*docs.Document
}

var _ = backendgoogle.AuthManager{}

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

func ReadGoogleDocHandler(ctx context.Context, rawArgs []byte, auth *backendgoogle.AuthManager) (any, error) {
	var read_args ReadGoogleDocsArgs
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
		Title: doc_id,
		Document: document,
	}, nil
}
