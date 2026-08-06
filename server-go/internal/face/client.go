package face

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client calls the Python face-service over HTTP. Embeddings are opaque bytes
// to Go — only the face-service interprets them.
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 30 * time.Second}}
}

// Enabled reports whether a face-service URL is configured.
func (c *Client) Enabled() bool { return c.baseURL != "" }

type EnrollResult struct {
	Embedding []byte
	Liveness  float64
}

func (c *Client) Enroll(ctx context.Context, image []byte) (*EnrollResult, error) {
	var out struct {
		Embedding     string  `json:"embedding"`
		LivenessScore float64 `json:"livenessScore"`
	}
	if err := c.post(ctx, "/enroll", map[string]string{"image": base64.StdEncoding.EncodeToString(image)}, &out); err != nil {
		return nil, err
	}
	emb, err := base64.StdEncoding.DecodeString(out.Embedding)
	if err != nil {
		return nil, fmt.Errorf("decode embedding: %w", err)
	}
	return &EnrollResult{Embedding: emb, Liveness: out.LivenessScore}, nil
}

type VerifyResult struct {
	Match    float64
	Liveness float64
	Passed   bool
}

func (c *Client) Verify(ctx context.Context, image, embedding []byte) (*VerifyResult, error) {
	body := map[string]string{
		"image":     base64.StdEncoding.EncodeToString(image),
		"embedding": base64.StdEncoding.EncodeToString(embedding),
	}
	var out struct {
		MatchScore    float64 `json:"matchScore"`
		LivenessScore float64 `json:"livenessScore"`
		Passed        bool    `json:"passed"`
	}
	if err := c.post(ctx, "/verify", body, &out); err != nil {
		return nil, err
	}
	return &VerifyResult{Match: out.MatchScore, Liveness: out.LivenessScore, Passed: out.Passed}, nil
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("face-service unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("face-service %s returned %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
