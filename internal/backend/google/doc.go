package google

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"encoding/json"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

type DocService struct {
	ctx context.Context 
	url string
	auth *AuthManager

}

func NewDocService(ctx context.Context, url string, auth *AuthManager) *DocService {
	return &DocService{
		ctx: ctx,
		url: url,
		auth: auth,
	}
}

func (ds *DocService ) GetDoc() (*docs.Document, error) {
	docId, err := ds.extractDocID(ds.url)

	httpClient, err := ds.auth.HTTPClient(ds.ctx)
	googleDocsHandler, err := docs.NewService(ds.ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	
	document, err := googleDocsHandler.Documents.Get(docId).Do()
	if err != nil {
		return nil, err
	}

	doc_bytes, _ := json.MarshalIndent(document, "", " ")
	fmt.Printf(string(doc_bytes))
	return document, nil
}


func (ds *DocService) extractDocID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("document url is empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid document url: %w", err)
	}

	if !strings.Contains(u.Host, "docs.google.com") {
		return "", fmt.Errorf("not a Google doc link: %s", raw)
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 4 || parts[1] != "document" || parts[2] != "d" || parts[3] == "" {
		return "", fmt.Errorf("could not extract doc id from url: %s", raw)
	}

	return parts[3], nil
}
