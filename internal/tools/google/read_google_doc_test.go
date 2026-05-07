package google

import (
	"testing"
	"context"
	"net/http"
)

type fakeAuthClient struct {
	client *http.Client
	err error
}

func (f fakeAuthClient) HTTPClient(ctx context.Context) (*http.Client, error) {
	return f.client, f.err
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
               