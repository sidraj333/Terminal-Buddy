package google

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeAuthClient struct {
	client *http.Client
	err    error
}

func (f fakeAuthClient) HTTPClient(ctx context.Context) (*http.Client, error) {
	return f.client, f.err
}

type fakeSuccessfulDocReadResponse struct{}

func (f fakeSuccessfulDocReadResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"title":"Test Doc"}`)),
		Request:    req,
	}, nil
}

func TestReadGoogleDocHandler_MissingDocumentURLReturnsError(t *testing.T) {
	rawArgs := []byte(`{}`)
	
	mockAuthClient := fakeAuthClient{
		client: &http.Client{},
		err: nil,

	}

	_, err := ReadGoogleDocHandler(context.Background(), rawArgs, mockAuthClient)
	if err == nil {
		t.Fatal("expected missing document_url to return an error")
	}
}

func TestReadGoogleDocHandler_HttpClientError(t *testing.T) {
	rawArgs := []byte(`{"document_url":"https://docs.google.com/document/d/test-doc-id/edit"}`)

	mockAuthClient := fakeAuthClient{
		client: nil,
		err:    fmt.Errorf("Error creating http client"),
	}

	_, err := ReadGoogleDocHandler(context.Background(), rawArgs, mockAuthClient)
	if err == nil {
		t.Fatal("expected to throw error when http client error occurs")
	}
}

func TestReadGoogleDocHandler_HappyPath(t *testing.T) {
	rawArgs := []byte(`{"document_url":"https://docs.google.com/document/d/test-doc-id/edit"}`)

	mockAuthClient := fakeAuthClient{
		client: &http.Client{
			Transport: fakeSuccessfulDocReadResponse{},
		},
		err: nil,
	}

	result, err := ReadGoogleDocHandler(context.Background(), rawArgs, mockAuthClient)
	if err != nil {
		t.Fatalf("expected happy path to succeed, got %v", err)
	}

	readResult, ok := result.(ReadGoogleDocResult)
	if !ok {
		t.Fatalf("expected result type %T, got %T", ReadGoogleDocResult{}, result)
	}

	if readResult.DocumentURL != "https://docs.google.com/document/d/test-doc-id/edit" {
		t.Fatalf("expected document URL %q, got %q", "https://docs.google.com/document/d/test-doc-id/edit", readResult.DocumentURL)
	}

	if readResult.Document == nil {
		t.Fatal("expected document to be populated")
	}

	if readResult.Title != "Test Doc" {
		t.Fatalf("expected title %q, got %q", "Test Doc", readResult.Title)
	}

}
               
