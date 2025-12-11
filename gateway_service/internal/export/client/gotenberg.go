package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type GotenbergClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewGotenbergClient(url string, timeoutSeconds int) *GotenbergClient {
	return &GotenbergClient{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (c *GotenbergClient) ConvertHTMLToPDF(ctx context.Context, html []byte, css []byte) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	htmlPart, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to create html form file: %w", err)
	}
	if _, err := htmlPart.Write(html); err != nil {
		return nil, fmt.Errorf("failed to write html: %w", err)
	}

	if len(css) > 0 {
		cssPart, err := writer.CreateFormFile("files", "style.css")
		if err != nil {
			return nil, fmt.Errorf("failed to create css form file: %w", err)
		}
		if _, err := cssPart.Write(css); err != nil {
			return nil, fmt.Errorf("failed to write css: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/forms/chromium/convert/html", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to gotenberg: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg returned status %d: %s", resp.StatusCode, string(respBody))
	}

	pdf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read pdf response: %w", err)
	}

	return pdf, nil
}

func (c *GotenbergClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gotenberg health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gotenberg unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
