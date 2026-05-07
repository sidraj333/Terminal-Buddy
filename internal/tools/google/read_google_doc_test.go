package google

import (
	"testing"
	"context"
)

func TestReadGoogleDocHandler_MissingDocumentURLReturnsError(t *testing.T) {
	rawArgs := []byte(`{}`)

	_, err := ReadGoogleDocHandler(context.Background(), rawArgs)
	if err == nil {
		t.Fatal("expected missing document_url to return an error")
	}
}
               