package google

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	gdocs "google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

type DocContext struct {
	DocID  string     `json:"doc_id"`
	URL    string     `json:"url"`
	Title  string     `json:"title"`
	Blocks []DocBlock `json:"blocks"`
}

type DocBlock struct {
	Type string `json:"type"` // paragraph | heading | table | unknown
	Text string `json:"text"`
}

type DocService struct {
	client *gdocs.Service
}

func NewDocsService(ctx context.Context, auth *AuthManager) (*DocService, error) {
	httpClient, err := auth.HTTPClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating authenticated http client: %w", err)
	}

	svc, err := gdocs.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("error creating docs service: %w", err)
	}

	return &DocService{client: svc}, nil
}

// ReadDocumentContextFromURL accepts a full Docs URL and returns structured context.
func (s *DocService) ReadDocumentContextFromURL(ctx context.Context, docURL string) (*DocContext, error) {
	docID, err := extractDocID(docURL)
	if err != nil {
		return nil, err
	}

	doc, err := s.client.Documents.Get(docID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("docs get %q: %w", docID, err)
	}

	out := &DocContext{
		DocID: docID,
		URL:   docURL,
		Title: doc.Title,
	}

	if doc.Body == nil {
		return out, nil
	}

	for _, c := range doc.Body.Content {
		if c == nil {
			continue
		}

		if c.Paragraph != nil {
			text := paragraphText(c.Paragraph)
			if strings.TrimSpace(text) == "" {
				continue
			}

			blockType := "paragraph"
			if c.Paragraph.ParagraphStyle != nil && c.Paragraph.ParagraphStyle.NamedStyleType != "" {
				ns := c.Paragraph.ParagraphStyle.NamedStyleType
				if strings.HasPrefix(ns, "HEADING_") {
					blockType = "heading"
				}
			}

			out.Blocks = append(out.Blocks, DocBlock{
				Type: blockType,
				Text: strings.TrimSpace(text),
			})
			continue
		}

		if c.Table != nil {
			out.Blocks = append(out.Blocks, DocBlock{
				Type: "table",
				Text: "[table content omitted]",
			})
			continue
		}

		out.Blocks = append(out.Blocks, DocBlock{
			Type: "unknown",
			Text: "[unsupported block type]",
		})
	}

	return out, nil
}

func paragraphText(p *gdocs.Paragraph) string {
	if p == nil {
		return ""
	}

	var b strings.Builder
	for _, e := range p.Elements {
		if e == nil || e.TextRun == nil {
			continue
		}
		b.WriteString(e.TextRun.Content)
	}

	return b.String()
}

func extractDocID(raw string) (string, error) {
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
