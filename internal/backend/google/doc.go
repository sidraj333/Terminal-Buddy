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
	document *docs.Document
	
}

func NewDocService(ctx context.Context, url string, auth *AuthManager) *DocService {
	return &DocService{
		ctx: ctx,
		url: url,
		auth: auth,
		document: nil,
	}
}

func (ds *DocService) Type() string { return "doc" } //satifies the source type interface used in main.go

func (ds *DocService ) Read()  error {
	docId, err := ds.extractDocID(ds.url)

	httpClient, err := ds.auth.HTTPClient(ds.ctx)
	googleDocsHandler, err := docs.NewService(ds.ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return err
	}
	
	document, err := googleDocsHandler.Documents.Get(docId).Do()
	if err != nil {
		return err
	}
	ds.document = document

	return nil
	
}

func (ds *DocService) Write() error {return nil}


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
