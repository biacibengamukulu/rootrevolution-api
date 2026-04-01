package dropbox

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"rootrevolution-api/config"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type ListEntry struct {
	Tag            string `json:".tag"`
	Name           string `json:"name"`
	PathLower      string `json:"path_lower"`
	PathDisplay    string `json:"path_display"`
	ID             string `json:"id"`
	Rev            string `json:"rev"`
	ClientModified string `json:"client_modified"`
	ServerModified string `json:"server_modified"`
	Size           int    `json:"size"`
}

type ListResponse struct {
	Entries []ListEntry `json:"entries"`
	Cursor  string      `json:"cursor"`
	HasMore bool        `json:"has_more"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.Dropbox.BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsBase64Image detects if a string is a base64-encoded image
func IsBase64Image(s string) bool {
	if strings.HasPrefix(s, "data:image/") {
		return true
	}
	// Check if it's a raw base64 string (long, no http prefix, not a URL)
	if !strings.HasPrefix(s, "http") && len(s) > 200 {
		_, err := base64.StdEncoding.DecodeString(s)
		return err == nil
	}
	return false
}

// UploadBase64Image decodes a base64 image, uploads to Dropbox, and returns the stream URL
func (c *Client) UploadBase64Image(base64Data, productID, filename string) (string, error) {
	var imgData []byte
	var ext string

	if strings.HasPrefix(base64Data, "data:image/") {
		// Parse data URI: data:image/png;base64,<data>
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid data URI format")
		}

		// Extract extension from MIME type
		mimeType := strings.TrimPrefix(strings.Split(parts[0], ";")[0], "data:")
		ext = mimeTypeToExt(mimeType)

		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("decoding base64 image: %w", err)
		}
		imgData = decoded
	} else {
		// Raw base64
		decoded, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", fmt.Errorf("decoding base64: %w", err)
		}
		imgData = decoded
		ext = ".jpg"
	}

	if filename == "" {
		filename = fmt.Sprintf("product_%s_%d%s", productID, time.Now().Unix(), ext)
	} else if filepath.Ext(filename) == "" {
		filename = filename + ext
	}

	dropboxPath := fmt.Sprintf("rootrevolution/products/%s/%s", productID, filename)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(imgData)); err != nil {
		return "", fmt.Errorf("writing file data: %w", err)
	}
	writer.Close()

	uploadURL := fmt.Sprintf("%s/upload?path=%s", c.baseURL, url.QueryEscape(dropboxPath))
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return "", fmt.Errorf("creating upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("uploading to dropbox: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dropbox upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Get the file rev by listing the uploaded path folder
	rev, err := c.getFileRev(fmt.Sprintf("rootrevolution/products/%s", productID), filename)
	if err != nil {
		// Return a fallback path-based reference if we can't get rev
		return fmt.Sprintf("%s/stream/%s", c.baseURL, dropboxPath), nil
	}

	return fmt.Sprintf("%s/stream/%s", c.baseURL, rev), nil
}

func (c *Client) getFileRev(folderPath, filename string) (string, error) {
	listURL := fmt.Sprintf("%s/list?path=%s", c.baseURL, url.QueryEscape(folderPath))
	resp, err := c.httpClient.Get(listURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return "", err
	}

	lowerFilename := strings.ToLower(filename)
	for _, entry := range listResp.Entries {
		if strings.ToLower(entry.Name) == lowerFilename && entry.Rev != "" {
			return entry.Rev, nil
		}
	}

	return "", fmt.Errorf("file %s not found in listing", filename)
}

func mimeTypeToExt(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
