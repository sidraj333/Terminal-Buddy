package google

import (
	"context"
	"encoding/json"
	"fmt"
)

type ReadGoogleDocsArgs struct {
	DocumentURL string `json:"document_url"`
}

type ReadGoogleDocResult struct {
	DocumentURL string `json:"document_url"`
	Title 		string `json:"title"`
	Text		string `json:"text"`
}

func ReadGoogleDocHandler(ctx context.Context, rawArgs []byte) (any, error) {
	var read_args ReadGoogleDocsArgs
	if err := json.Unmarshal(rawArgs, &read_args); err != nil {
		return ReadGoogleDocResult{}, err
	}
	
	if read_args.DocumentURL == "" {
		return ReadGoogleDocResult{}, fmt.Errorf("expected document url as a parameter")
	}
	
	return ReadGoogleDocResult{}, nil
}