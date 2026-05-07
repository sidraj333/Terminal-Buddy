package google

  import (
        "context"
        "net/http"
  )

  type HTTPClientProvider interface {
        HTTPClient(ctx context.Context) (*http.Client, error)
  }