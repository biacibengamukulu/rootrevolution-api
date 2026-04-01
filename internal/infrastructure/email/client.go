package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"rootrevolution-api/config"
)

type Client struct {
	baseURL    string
	org        string
	httpClient *http.Client
}

type EmailPayload struct {
	Org     string `json:"org"`
	From    string `json:"from"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.Email.BaseURL,
		org:     cfg.Email.Org,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) Send(recipientEmail, recipientName, subject, htmlMessage string) error {
	payload := EmailPayload{
		Org:     c.org,
		From:    "Root Revolution System",
		Name:    "Product Catalogue Notification",
		Email:   recipientEmail,
		Subject: subject,
		Message: htmlMessage,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling email payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating email request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending email: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) SendProductAuthorizationEmail(ownerEmail, ownerName, authLink, action, productName, requestedBy string) error {
	actionLabel := "updated"
	if action == "create" {
		actionLabel = "created"
	} else if action == "delete" {
		actionLabel = "deleted"
	}

	subject := fmt.Sprintf("Authorization Required: Product %q %s", productName, actionLabel)
	message := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background: #1a1a2e; color: #ffffff; padding: 20px; border-radius: 8px 8px 0 0;">
    <h1 style="margin: 0; font-size: 24px;">Root Revolution</h1>
    <p style="margin: 5px 0 0 0; color: #a0a0b0;">Product Catalogue Management</p>
  </div>
  <div style="background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; border: 1px solid #e0e0e0;">
    <h2 style="color: #1a1a2e;">Authorization Required</h2>
    <p>Hello <strong>%s</strong>,</p>
    <p>A product catalogue change has been requested and requires your authorization:</p>
    <table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
      <tr style="background: #eef2ff;">
        <td style="padding: 10px; border: 1px solid #ddd; font-weight: bold;">Product</td>
        <td style="padding: 10px; border: 1px solid #ddd;">%s</td>
      </tr>
      <tr>
        <td style="padding: 10px; border: 1px solid #ddd; font-weight: bold;">Action</td>
        <td style="padding: 10px; border: 1px solid #ddd; text-transform: capitalize;">%s</td>
      </tr>
      <tr style="background: #eef2ff;">
        <td style="padding: 10px; border: 1px solid #ddd; font-weight: bold;">Requested By</td>
        <td style="padding: 10px; border: 1px solid #ddd;">%s</td>
      </tr>
    </table>
    <p>To authorize this change, click the button below. This link expires in <strong>24 hours</strong>.</p>
    <div style="text-align: center; margin: 30px 0;">
      <a href="%s" style="background: #4f46e5; color: #ffffff; padding: 14px 30px; text-decoration: none; border-radius: 6px; font-size: 16px; font-weight: bold; display: inline-block;">
        Authorize Change
      </a>
    </div>
    <p style="color: #666; font-size: 13px;">If you did not request this change or do not recognize it, please ignore this email. No changes will be made without your authorization.</p>
    <hr style="border: none; border-top: 1px solid #e0e0e0; margin: 20px 0;">
    <p style="color: #999; font-size: 12px;">Root Revolution Product Management System</p>
  </div>
</body>
</html>`,
		ownerName, productName, actionLabel, requestedBy, authLink,
	)

	return c.Send(ownerEmail, ownerName, subject, message)
}
