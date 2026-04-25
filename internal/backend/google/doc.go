package google

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"encoding/json"
	"net/http"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
	"bytes"
	"io"
)

type DocService struct {
	ctx context.Context 
	url string
	auth *AuthManager
	document *docs.Document
	doc_id string
	source_type string
	
}

func NewDocService(ctx context.Context, sourceUrl string, auth *AuthManager) (*DocService, error) {
		raw := strings.TrimSpace(sourceUrl)
		if raw == "" {
			return nil, errors.New("document url is empty")
		}

		u, err := url.Parse(sourceUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid document url: %w", err)
		}

		if !strings.Contains(u.Host, "docs.google.com") {
			return nil, fmt.Errorf("not a valid Google doc link: %s", raw)
		}

		parts := strings.Split(u.Path, "/")
		if len(parts) < 4 || parts[1] != "document" || parts[2] != "d" || parts[3] == "" {
			return nil, fmt.Errorf("could not extract doc id from url: %s", raw)
		}
		doc_id := parts[3]
		
	
		


	return &DocService{
		ctx: ctx,
		url: sourceUrl,
		auth: auth,
		document: nil,
		doc_id: doc_id,
		source_type: "doc",
	}, nil
}

func (ds *DocService) Type() string { return "doc" } //satifies the source type interface used in main.go

func (ds *DocService ) Read()  error {

	httpClient, err := ds.auth.HTTPClient(ds.ctx)
	googleDocsHandler, err := docs.NewService(ds.ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return err
	}
	
	document, err := googleDocsHandler.Documents.Get(ds.doc_id).Do()
	if err != nil {
		return err
	}
	ds.document = document

	return nil
	
}

func (ds *DocService) Write() error {return nil}


func (ds *DocService) Ask(question string) (string, error) {
	//POST request to chanlder_.go with document information and the question
	ds.Read()

	marshalledDoc, err := json.Marshal(ds.document)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ds.ctx, http.MethodPost, "http://localhost:/8080", bytes.NewReader(marshalledDoc),)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	resp_bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return string(resp_bytes), nil

	
}