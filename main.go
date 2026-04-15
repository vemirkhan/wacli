// wacli - WhatsApp CLI tool
// Fork of steipete/wacli
// Sends WhatsApp messages via the WhatsApp Business API
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	baseURL        = "https://graph.facebook.com/v18.0"
	envToken       = "WA_TOKEN"
	envPhoneID     = "WA_PHONE_ID"
	envRecipient   = "WA_RECIPIENT"
)

// MessageRequest represents the WhatsApp API message payload
type MessageRequest struct {
	MessagingProduct string   `json:"messaging_product"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Text             *TextMsg `json:"text,omitempty"`
}

// TextMsg holds the text body for a text message
type TextMsg struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

func main() {
	var (
		token     string
		phoneID   string
		recipient string
	)

	rootCmd := &cobra.Command{
		Use:   "wacli [message]",
		Short: "wacli sends WhatsApp messages via the WhatsApp Business API",
		Long: `wacli is a command-line tool to send WhatsApp messages
using the Meta WhatsApp Business Cloud API.

Credentials can be provided via flags or environment variables:
  WA_TOKEN      - WhatsApp API access token
  WA_PHONE_ID   - Sender phone number ID
  WA_RECIPIENT  - Recipient phone number (with country code, e.g. 14155552671)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			message := strings.TrimSpace(args[0])
			if message == "" {
				return fmt.Errorf("message cannot be empty")
			}

			// Fall back to environment variables if flags not set
			if token == "" {
				token = os.Getenv(envToken)
			}
			if phoneID == "" {
				phoneID = os.Getenv(envPhoneID)
			}
			if recipient == "" {
				recipient = os.Getenv(envRecipient)
			}

			if token == "" {
				return fmt.Errorf("API token is required (--token or %s)", envToken)
			}
			if phoneID == "" {
				return fmt.Errorf("phone ID is required (--phone-id or %s)", envPhoneID)
			}
			if recipient == "" {
				return fmt.Errorf("recipient is required (--to or %s)", envRecipient)
			}

			return sendMessage(token, phoneID, recipient, message)
		},
	}

	rootCmd.Flags().StringVarP(&token, "token", "t", "", "WhatsApp API access token")
	rootCmd.Flags().StringVarP(&phoneID, "phone-id", "p", "", "Sender phone number ID")
	rootCmd.Flags().StringVarP(&recipient, "to", "r", "", "Recipient phone number (with country code)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// sendMessage sends a text message via the WhatsApp Business Cloud API
func sendMessage(token, phoneID, recipient, message string) error {
	payload := MessageRequest{
		MessagingProduct: "whatsapp",
		To:               recipient,
		Type:             "text",
		Text: &TextMsg{
			PreviewURL: false,
			Body:       message,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", baseURL, phoneID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error ( %d): %s", resp.StatusCode, string(respBody))
	}

	fmt.Println("n}
